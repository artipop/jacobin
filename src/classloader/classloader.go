/*
 * Jacobin VM - A Java virtual machine
 * Copyright (c) 2021-4 by the Jacobin authors. All rights reserved.
 * Licensed under Mozilla Public License 2.0 (MPL 2.0)
 */

package classloader

import (
	"errors"
	"fmt"
	"io/fs"
	"jacobin/excNames"
	"jacobin/globals"
	"jacobin/shutdown"
	"jacobin/stringPool"
	"jacobin/trace"
	"jacobin/types"
	"jacobin/util"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
)

// Classloader holds the parsed bytecode in classes, where they can be retrieved
// and moved to an execution role. Most of the comments and code presuppose some
// familiarity with the role of classloaders. More information can be found at:
// https://docs.oracle.com/javase/specs/jvms/se17/html/jvms-5.html#jvms-5.3
type Classloader struct {
	Name       string
	Parent     string
	ClassCount int
	Archives   map[string]*Archive // TODO: I think this should be moved to classpath when we make it a thing
}

// AppCL is the application classloader, which loads most of the app's classes
var AppCL Classloader

// BootstrapCL is the classloader that loads most of the standard libraries
var BootstrapCL Classloader

// ExtensionCL is the classloader typically used for loading custom agents
var ExtensionCL Classloader

// ParsedClass contains all the parsed fields
type ParsedClass struct {
	javaVersion    int
	className      string // name of class without path and without .class TODO: eventually remove
	classNameIndex uint32 // index into StringPool
	// superClass      string // name of superclass for this class TODO: eventually remove in favor of stringPool
	superClassIndex uint32 // index of into StringPool
	moduleName      string
	packageName     string
	interfaceCount  int      // number of interfaces this class implements
	interfaces      []uint32 // the interfaces this class implements, as indices into the string pool
	fieldCount      int      // number of fields in this class
	fields          []field
	methodCount     int
	methods         []method
	attribCount     int
	attributes      []attr
	sourceFile      string
	bootstrapCount  int // the number of bootstrap methods
	bootstraps      []bootstrapMethod

	deprecated bool

	// ---- constant pool data items ----
	cpCount        int       // count of constant pool entries
	cpIndex        []cpEntry // the constant pool index to entries
	classRefs      []uint32  // point to a stringPool index to a class name
	doubles        []float64
	dynamics       []dynamic
	fieldRefs      []fieldRefEntry
	floats         []float32
	intConsts      []int // 32-bit int containing the actual int value
	interfaceRefs  []interfaceRefEntry
	invokeDynamics []invokeDynamic
	longConsts     []int64
	methodHandles  []methodHandleEntry
	methodRefs     []methodRefEntry
	methodTypes    []int
	nameAndTypes   []nameAndTypeEntry
	stringRefs     []stringConstantEntry // integer index into utf8Refs
	utf8Refs       []utf8Entry

	// ---- access flags items ----
	accessFlags       int // the following booleans interpret the access flags
	classIsPublic     bool
	classIsFinal      bool
	classIsSuper      bool
	classIsInterface  bool
	classIsAbstract   bool
	classIsSynthetic  bool
	classIsAnnotation bool
	classIsEnum       bool
	classIsModule     bool

	// ---- other attributes ----

}

// the fields defined in the class
type field struct {
	accessFlags int
	isStatic    bool
	name        int         // index of the UTF-8 entry in the CP
	description int         // index of the UTF-8 entry in the CP
	constValue  interface{} // the constant value if any was defined
	attributes  []attr
}

// the methods of the class, including the constructors
type method struct {
	accessFlags int
	name        int // index of the UTF-8 entry in the CP
	description int // index of the UTF-8 entry in the CP
	codeAttr    codeAttrib
	attributes  []attr
	exceptions  []uint32 // indexes into constant pool,
	// pointing to names of exception classes this method is knownto throw
	parameters []paramAttrib
	deprecated bool // is the method deprecated?
}

type codeAttrib struct {
	maxStack        int
	maxLocals       int
	code            []byte
	exceptions      []exception // exception entries for this method
	attributes      []attr      // the code attributes has its own sub-attributes(!)
	sourceLineTable *[]BytecodeToSourceLine
}

// the MethodParameters method attribute
type paramAttrib struct {
	name        string // string, rather than index into utf8Refs b/c the name could be ""
	accessFlags int
}

// the structure of many attributes (field, class, etc.) The content is just the raw bytes.
type attr struct {
	attrName    int    // index of the UTF-8 entry in the CP
	attrSize    int    // length of the following array of raw bytes
	attrContent []byte // the raw data of the attribute
}

// the exception-related data for each exception in the Code attribute of a given method
type exception struct {
	startPc   int // first instruction covered by this exception (pc = program counter)
	endPc     int // the last instruction covered by this exception
	handlerPc int // the place in the method code that has the exception instructions
	catchType int // the type of exception, index to CP, which must point a ClassFref entry
}

// the bootstrap methods, specified in the bootstrap class attribute
type bootstrapMethod struct {
	methodRef int   // index pointing to a MethodHandle
	args      []int // arguments: indexes to loadable arguments from the CP
}

var ClassesLock = sync.RWMutex{}

// cfe = class format error, which is the error thrown by the parser for most
// of the errors arising from malformed bytecode. Prints out file and line# where
// the call to cfe() occurred.
func cfe(msg string) error {
	errMsg := "Class Format Error: " + msg

	// get the filename and line# of the function where the error occurred
	// implementation note: Caller(0) would be this function. (1) is the
	// previous function on the stack (so, the one calling this error routine)
	// To traverse all the way back to the start of the program, set up a loop
	// and exit when ok is no longer true.
	pc, _, _, ok := runtime.Caller(1)
	if ok {
		fn := runtime.FuncForPC(pc)
		fileName, fileLine := fn.FileLine(pc)
		errMsg = errMsg + "\n  detected by file: " + filepath.Base(fileName) +
			", line: " + strconv.Itoa(fileLine)
	}
	trace.Error(errMsg)
	return errors.New(errMsg)
}

func CFE(msg string) error { return cfe(msg) }

// LoadBaseClasses loads a basic set of classes that are found in
// the JAVA_HOME/jmods/java.base.jmod zip file.
// In Java 17.0.7, there are currently a total of 6401 embedded classes in java.base.jmod.
// Based on the lib/classlist member in java.base.jmod, only 1402 class files are actually loaded by this function.
func LoadBaseClasses() {
	global := globals.GetGlobalRef()
	jmodFilePath := global.JavaHome + string(os.PathSeparator) + "jmods" + string(os.PathSeparator) + "java.base.jmod"

	err := WalkBaseJmod()
	if err != nil {
		errMsg := fmt.Sprintf("LoadBaseClasses: Error loading jmod file classes %s, err: %v", jmodFilePath, err)
		trace.Error(errMsg)
		shutdown.Exit(shutdown.JVM_EXCEPTION)
	}

	if globals.TraceCloadi {
		infoMsg := fmt.Sprintf("LoadBaseClasses: Bootstrap classes from %s have been loaded", jmodFilePath)
		trace.Trace(infoMsg)
	}

}

// walk the directory and load every file (which is known to be a class)
// TODO: test work on JAR files to determine whether this function is still used
func walk(s string, d fs.DirEntry, err error) error {
	if err != nil {
		return err
	}
	if !d.IsDir() && strings.HasSuffix(s, ".class") {
		// Error is discarded b/c it's not clear yet a given class is needed.
		_, _, _ = LoadClassFromFile(BootstrapCL, s)
	}
	return nil
}

// LoadFromLoaderChannel receives a name of a class to load in /java/lang/String format,
// determines the classloader, checks if the class is already loaded, and loads it if not.
//
// Note: per JACOBIN-327, this parallel processing has been temporarily
// removed--it's not called by any function. It's likely to be reinstated later.
func LoadFromLoaderChannel(LoaderChannel <-chan string) {
	for name := range LoaderChannel {
		present := MethAreaFetch(name)
		if present != nil { // if the class is already loaded, skip it
			continue
		}

		// add entry to the method area, indicating initialization of the load of this class
		eKI := Klass{
			Status: 'I', // I = initializing the load
			Loader: "",
			Data:   nil,
		}
		MethAreaInsert(name, &eKI)
		err := LoadClassFromNameOnly(util.ConvertToPlatformPathSeparators(name))
		if err != nil {
			shutdown.Exit(shutdown.JVM_EXCEPTION)
		}
	}
	globals.LoaderWg.Done()
}

// LoadClassFromNameOnly loads a class from name in java/lang/Class format
// It also loads the superclass of any class it loads.
func LoadClassFromNameOnly(name string) error {
	var err error
	className := name

loadAclass:

	if className == "" {
		errMsg := "LoadClassFromNameOnly(): null class name is invalid"
		trace.Error(errMsg)
		return errors.New(errMsg)
	}

	// get the jmod file name for this class. We'll use the jmod file to
	// get the .class file for this class.
	jmodFileName := JmodMapFetch(className)

	if strings.HasSuffix(className, ";") {
		errMsg := fmt.Sprintf("LoadClassFromNameOnly: invalid class name: %s", className)
		trace.Error(errMsg)
		return errors.New(errMsg)
	}

	// Load class from a jmod?
	if jmodFileName != "" {
		if globals.TraceClass {
			trace.Trace("LoadClassFromNameOnly: Load " + className + " from jmod " + jmodFileName)
		}
		classBytes, err := GetClassBytes(jmodFileName, className)
		if err != nil {
			errMsg := "LoadClassFromNameOnly: GetClassBytes className=" + className + " from jmodFileName=" + jmodFileName + " failed, err: " + err.Error()
			trace.Error(errMsg)
		}
		_, _, err = loadClassFromBytes(AppCL, className, classBytes)
		return err
	}

	// Load class from a jar file?
	if len(globals.GetGlobalRef().StartingJar) > 0 {
		validName := util.ConvertToPlatformPathSeparators(className)
		if globals.TraceClass {
			trace.Trace("LoadClassFromNameOnly: LoadClassFromJar " + validName)
		}
		_, _, err = LoadClassFromJar(AppCL, validName, globals.GetGlobalRef().StartingJar)
		if err != nil {
			errMsg := "LoadClassFromNameOnly: LoadClassFromJar " + validName + " failed, err: " + err.Error()
			trace.Error(errMsg)
		}
		return err
	}

	// Loading from a local file system class
	// TODO: classpath
	validName := util.ConvertToPlatformPathSeparators(className)
	if globals.TraceClass {
		trace.Trace("LoadClassFromNameOnly: Loaded class from file " + validName)
	}
	_, superclassIndex, err := LoadClassFromFile(AppCL, validName)
	if err != nil {
		errMsg := fmt.Sprintf("LoadClassFromNameOnly for %s failed, err: %v", className, err)
		globals.GetGlobalRef().FuncThrowException(excNames.ClassNotFoundException, errMsg)
		return errors.New(errMsg) // return for tests only
	}

	// load any superclass in a recursive fashion
	if superclassIndex != types.ObjectPoolStringIndex { // don't load if it's java/lang/Object
		className = *stringPool.GetStringPointer(superclassIndex)
		goto loadAclass
	}
	return err
}

// LoadClassFromFile first canonicalizes the filename, and reads
// the indicated file, and runs it through the classloader.
func LoadClassFromFile(cl Classloader, fname string) (uint32, uint32, error) {
	var filename string
	if !strings.HasSuffix(fname, ".class") {
		filename = fname + ".class"
	} else {
		filename = fname
	}
	if filename == ".class" || strings.HasSuffix(filename, ";.class") {
		errMsg := "LoadClassFromFile: class name" + fname + " is invalid"
		trace.Error(errMsg)
		debug.PrintStack()
		return types.InvalidStringIndex, types.InvalidStringIndex, errors.New(errMsg)
	}
	rawBytes, err := os.ReadFile(filename)
	if err != nil {
		errMsg := fmt.Sprintf("LoadClassFromFile for %s failed", filename)
		globals.GetGlobalRef().FuncThrowException(excNames.ClassNotFoundException, errMsg)
		return types.InvalidStringIndex, types.InvalidStringIndex, errors.New(errMsg) // return for tests only
	}
	if globals.TraceClass {
		trace.Trace("LoadClassFromFile: File " + fname + " was read")
	}

	return loadClassFromBytes(cl, filename, rawBytes)
}

func getJarFile(cl Classloader, jarFileName string) (*Archive, error) {
	archive, exists := cl.Archives[jarFileName]

	if exists {
		return archive, nil
	}

	jar, err := NewJarFile(jarFileName)

	if err != nil {
		return nil, err
	}

	cl.Archives[jarFileName] = jar

	return jar, nil
}

func GetMainClassFromJar(cl Classloader, jarFileName string) (string, error) {
	jar, err := getJarFile(cl, jarFileName)

	if err != nil {
		return "", err
	}

	return jar.getMainClass(), nil
}

func LoadClassFromJar(cl Classloader, filename string, jarFileName string) (uint32, uint32, error) {
	jar, err := getJarFile(cl, jarFileName)

	if err != nil {
		return types.InvalidStringIndex, types.InvalidStringIndex, err
	}

	result, err := jar.loadClass(filename)

	if err != nil {
		return types.InvalidStringIndex, types.InvalidStringIndex, err
	}

	if !result.Success {
		return types.InvalidStringIndex, types.InvalidStringIndex,
			fmt.Errorf("unable to find file %s in JAR file %s", filename, jarFileName)
	}

	return ParseAndPostClass(&cl, filename, *result.Data)
}

func loadClassFromBytes(cl Classloader, filename string, rawBytes []byte) (uint32, uint32, error) {
	return ParseAndPostClass(&cl, filename, rawBytes)
}

// ParseAndPostClass parses a class, presented as a slice of bytes, and
// if no errors occurred, posts/loads it to the method area.
func ParseAndPostClass(cl *Classloader, filename string, rawBytes []byte) (uint32, uint32, error) {

	if globals.TraceClass {
		trace.Trace("ParseAndPostClass: File " + filename + " to be processed")
	}
	fullyParsedClass, err := parse(rawBytes)
	if err != nil {
		trace.Error("ParseAndPostClass: file " + filename + ", err: " + err.Error())
		return types.InvalidStringIndex, types.InvalidStringIndex, fmt.Errorf("parsing error")
	}

	// format check the class
	if formatCheckClass(&fullyParsedClass) != nil {
		trace.Error("ParseAndPostClass: format-checking " + filename)
		return types.InvalidStringIndex, types.InvalidStringIndex, fmt.Errorf("format-checking error")
	}
	if globals.TraceClass {
		trace.Trace("Class " + fullyParsedClass.className + " has been format-checked.")
	}

	// prepare the class for posting
	classToPost := convertToPostableClass(&fullyParsedClass)
	eKF := Klass{
		Status: 'F', // F = format-checked
		Loader: cl.Name,
		Data:   &classToPost,
	}
	MethAreaInsert(fullyParsedClass.className, &eKF)

	// record the class in the classloader
	ClassesLock.Lock()
	cl.ClassCount += 1
	ClassesLock.Unlock()
	if globals.TraceClass {
		trace.Trace("ParseAndPostClass: File " + filename + " fully processed")
	}

	return fullyParsedClass.classNameIndex, fullyParsedClass.superClassIndex, nil
}

// load the parsed class into a form suitable for posting to the method area (which is
// classloader.MethArea). This mostly involves copying the data, converting most indexes
// to uint16 and removing some fields we needed in parsing, but which are no longer required.
//
// As of JACOBIN-575, methods are no longer included in the MethArea, but are all loaded into
// the JVM-wide MTable (classloader.mTable).
func convertToPostableClass(fullyParsedClass *ParsedClass) ClData {

	kd := ClData{}

	kd.Name = fullyParsedClass.className // eventually to be deleted in favor of class index
	kd.NameIndex = fullyParsedClass.classNameIndex
	// kd.Superclass = fullyParsedClass.superClass // eventually to be delete in favor of class index
	kd.SuperclassIndex = fullyParsedClass.superClassIndex

	kd.Module = fullyParsedClass.moduleName
	kd.Pkg = fullyParsedClass.packageName
	for i := 0; i < len(fullyParsedClass.interfaces); i++ {
		kd.Interfaces = append(kd.Interfaces, uint16(fullyParsedClass.interfaces[i]))
	}
	if len(fullyParsedClass.fields) > 0 {
		for i := 0; i < len(fullyParsedClass.fields); i++ {
			kdf := Field{}
			kdf.Name = uint16(fullyParsedClass.fields[i].name)
			kdf.Desc = uint16(fullyParsedClass.fields[i].description)
			kdf.IsStatic = fullyParsedClass.fields[i].isStatic
			if len(fullyParsedClass.fields[i].attributes) > 0 {
				for j := 0; j < len(fullyParsedClass.fields[i].attributes); j++ {
					kdfa := Attr{}
					kdfa.AttrName = uint16(fullyParsedClass.fields[i].attributes[j].attrName)
					kdfa.AttrSize = fullyParsedClass.fields[i].attributes[j].attrSize
					kdfa.AttrContent = fullyParsedClass.fields[i].attributes[j].attrContent
					kdf.Attributes = append(kdf.Attributes, kdfa)
				}
			}
			kd.Fields = append(kd.Fields, kdf)
		}
	}

	kd.MethodTable = make(map[string]*Method)
	if len(fullyParsedClass.methods) > 0 {
		for i := 0; i < len(fullyParsedClass.methods); i++ {
			jmeth := JmEntry{}
			jmeth.CodeAttr = CodeAttrib{}

			kdm := Method{}
			kdm.Name = uint16(fullyParsedClass.methods[i].name)
			methName := fullyParsedClass.utf8Refs[int(kdm.Name)].content
			kdm.Desc = uint16(fullyParsedClass.methods[i].description)
			methDesc := fullyParsedClass.utf8Refs[int(kdm.Desc)].content

			kdm.AccessFlags = fullyParsedClass.methods[i].accessFlags
			jmeth.AccessFlags = fullyParsedClass.methods[i].accessFlags

			kdm.CodeAttr.MaxStack = fullyParsedClass.methods[i].codeAttr.maxStack
			jmeth.MaxStack = fullyParsedClass.methods[i].codeAttr.maxStack

			kdm.CodeAttr.MaxLocals = fullyParsedClass.methods[i].codeAttr.maxLocals
			jmeth.MaxLocals = fullyParsedClass.methods[i].codeAttr.maxLocals

			kdm.CodeAttr.Code = fullyParsedClass.methods[i].codeAttr.code
			jmeth.Code = fullyParsedClass.methods[i].codeAttr.code

			if len(fullyParsedClass.methods[i].codeAttr.exceptions) > 0 {
				for j := 0; j < len(fullyParsedClass.methods[i].codeAttr.exceptions); j++ {
					kdmce := CodeException{}
					kdmce.StartPc = fullyParsedClass.methods[i].codeAttr.exceptions[j].startPc
					kdmce.EndPc = fullyParsedClass.methods[i].codeAttr.exceptions[j].endPc
					kdmce.HandlerPc = fullyParsedClass.methods[i].codeAttr.exceptions[j].handlerPc
					kdmce.CatchType = uint16(fullyParsedClass.methods[i].codeAttr.exceptions[j].catchType)
					kdm.CodeAttr.Exceptions = append(kdm.CodeAttr.Exceptions, kdmce)
					jmeth.CodeAttr.Exceptions = append(jmeth.CodeAttr.Exceptions, kdmce)
				}
			}

			if len(fullyParsedClass.methods[i].codeAttr.attributes) > 0 {
				for m := 0; m < len(fullyParsedClass.methods[i].codeAttr.attributes); m++ {
					kdmca := Attr{}
					kdmca.AttrName = uint16(fullyParsedClass.methods[i].codeAttr.attributes[m].attrName)
					kdmca.AttrSize = fullyParsedClass.methods[i].codeAttr.attributes[m].attrSize
					kdmca.AttrContent = fullyParsedClass.methods[i].codeAttr.attributes[m].attrContent
					kdm.CodeAttr.Attributes = append(kdm.CodeAttr.Attributes, kdmca)
					jmeth.CodeAttr.Attributes = append(jmeth.CodeAttr.Attributes, kdmca)
				}
			}

			if fullyParsedClass.methods[i].codeAttr.sourceLineTable != nil {
				if len(*fullyParsedClass.methods[i].codeAttr.sourceLineTable) > 0 {
					jmeth.CodeAttr.BytecodeSourceMap = *fullyParsedClass.methods[i].codeAttr.sourceLineTable
				}
			} else {
				fullyParsedClass.methods[i].codeAttr.sourceLineTable = nil
			}

			if len(fullyParsedClass.methods[i].attributes) > 0 {
				for n := 0; n < len(fullyParsedClass.methods[i].attributes); n++ {
					kdma := Attr{
						AttrName:    uint16(fullyParsedClass.methods[i].attributes[n].attrName),
						AttrSize:    fullyParsedClass.methods[i].attributes[n].attrSize,
						AttrContent: fullyParsedClass.methods[i].attributes[n].attrContent,
					}
					kdm.Attributes = append(kdm.Attributes, kdma)
					jmeth.Attribs = append(jmeth.Attribs, kdma)
				}
			}

			// exceptions here are simply indexes into the CP, pointing to class references
			// for each exception that is declared for this method to throw. See:
			// https://docs.oracle.com/javase/specs/jvms/se17/html/jvms-4.html#jvms-4.7.5
			if len(fullyParsedClass.methods[i].exceptions) > 0 {
				for p := 0; p < len(fullyParsedClass.methods[i].exceptions); p++ {
					kdm.Exceptions =
						append(kdm.Exceptions, uint16(fullyParsedClass.methods[i].exceptions[p]))
				} // CURR
			}

			if len(fullyParsedClass.methods[i].parameters) > 0 {
				for q := 0; q < len(fullyParsedClass.methods[i].parameters); q++ {
					kdmp := ParamAttrib{
						Name:        fullyParsedClass.methods[i].parameters[q].name,
						AccessFlags: fullyParsedClass.methods[i].parameters[q].accessFlags,
					}
					kdm.Parameters = append(kdm.Parameters, kdmp)
					jmeth.params = append(jmeth.params, kdmp)
				}
			}
			kdm.Deprecated = fullyParsedClass.methods[i].deprecated
			jmeth.deprecated = fullyParsedClass.methods[i].deprecated

			methodTableKey := methName + methDesc
			kd.MethodTable[methodTableKey] = &kdm
		}
	} // end of methods processing

	_, clInitPresent := kd.MethodTable["<clinit>()V"]
	if clInitPresent {
		kd.ClInit = types.ClInitNotRun // there is a clinit, but it's not been run
	} else {
		kd.ClInit = types.NoClinit // there is no clinit
	}

	if len(fullyParsedClass.attributes) > 0 {
		for i := 0; i < len(fullyParsedClass.attributes); i++ {
			kda := Attr{
				AttrName:    uint16(fullyParsedClass.attributes[i].attrName),
				AttrSize:    fullyParsedClass.attributes[i].attrSize,
				AttrContent: fullyParsedClass.attributes[i].attrContent,
			}
			kd.Attributes = append(kd.Attributes, kda)
		}
	}
	kd.SourceFile = fullyParsedClass.sourceFile
	if len(fullyParsedClass.bootstraps) > 0 {
		for j := 0; j < len(fullyParsedClass.bootstraps); j++ {
			kdbs := BootstrapMethod{
				MethodRef: uint16(fullyParsedClass.bootstraps[j].methodRef),
				Args:      nil,
			}
			if len(fullyParsedClass.bootstraps[j].args) > 0 {
				for l := 0; l < len(fullyParsedClass.bootstraps[j].args); l++ {
					kdbs.Args = append(kdbs.Args, uint16(fullyParsedClass.bootstraps[j].args[l]))
				}
			}
			kd.Bootstraps = append(kd.Bootstraps, kdbs)
		}
	}
	kd.Access.ClassIsPublic = fullyParsedClass.classIsPublic
	kd.Access.ClassIsFinal = fullyParsedClass.classIsFinal
	kd.Access.ClassIsSuper = fullyParsedClass.classIsSuper
	kd.Access.ClassIsInterface = fullyParsedClass.classIsInterface
	kd.Access.ClassIsAbstract = fullyParsedClass.classIsAbstract
	kd.Access.ClassIsSynthetic = fullyParsedClass.classIsSynthetic
	kd.Access.ClassIsAnnotation = fullyParsedClass.classIsAnnotation
	kd.Access.ClassIsEnum = fullyParsedClass.classIsEnum
	kd.Access.ClassIsModule = fullyParsedClass.classIsModule

	// ---- loading the CP ----
	for i := 0; i < fullyParsedClass.cpCount; i++ {

		// most CP entries are brought over with minor changes (indexes are shortened to uint16, etc.);
		// however, stringRefs are converted to UTF-8 references here before being brought over.
		if fullyParsedClass.cpIndex[i].entryType == StringConst {
			whichStringConst := fullyParsedClass.cpIndex[i].slot
			cpIndexForUTF8 := fullyParsedClass.stringRefs[whichStringConst]
			cpE := CpEntry{
				Type: UTF8,
				Slot: uint16(fullyParsedClass.cpIndex[cpIndexForUTF8.index].slot),
			}
			kd.CP.CpIndex = append(kd.CP.CpIndex, cpE)
		} else {
			cpE := CpEntry{
				Type: uint16(fullyParsedClass.cpIndex[i].entryType),
				Slot: uint16(fullyParsedClass.cpIndex[i].slot),
			}
			kd.CP.CpIndex = append(kd.CP.CpIndex, cpE)
		}
	}

	if len(fullyParsedClass.classRefs) > 0 {
		for i := 0; i < len(fullyParsedClass.classRefs); i++ {
			kd.CP.ClassRefs = append(kd.CP.ClassRefs, fullyParsedClass.classRefs[i])
		}
	}

	if len(fullyParsedClass.doubles) > 0 {
		for i := 0; i < len(fullyParsedClass.doubles); i++ {
			kd.CP.Doubles = append(kd.CP.Doubles, fullyParsedClass.doubles[i])
		}
	}

	if len(fullyParsedClass.dynamics) > 0 {
		for i := 0; i < len(fullyParsedClass.dynamics); i++ {
			dyn := DynamicEntry{
				BootstrapIndex: uint16(fullyParsedClass.dynamics[i].bootstrapIndex),
				NameAndType:    uint16(fullyParsedClass.dynamics[i].nameAndType),
			}
			kd.CP.Dynamics = append(kd.CP.Dynamics, dyn)
		}
	}

	if len(fullyParsedClass.fieldRefs) > 0 {
		for i := 0; i < len(fullyParsedClass.fieldRefs); i++ {
			fr := FieldRefEntry{
				ClassIndex:  uint16(fullyParsedClass.fieldRefs[i].classIndex),
				NameAndType: uint16(fullyParsedClass.fieldRefs[i].nameAndTypeIndex),
			}
			kd.CP.FieldRefs = append(kd.CP.FieldRefs, fr)
		}
	}

	if len(fullyParsedClass.floats) > 0 {
		for i := 0; i < len(fullyParsedClass.floats); i++ {
			kd.CP.Floats = append(kd.CP.Floats, fullyParsedClass.floats[i])
		}
	}

	if len(fullyParsedClass.intConsts) > 0 {
		for i := 0; i < len(fullyParsedClass.intConsts); i++ {
			kd.CP.IntConsts = append(kd.CP.IntConsts, int32(fullyParsedClass.intConsts[i]))
		}
	}

	if len(fullyParsedClass.interfaceRefs) > 0 {
		for i := 0; i < len(fullyParsedClass.interfaceRefs); i++ {
			ir := InterfaceRefEntry{
				ClassIndex:  uint16(fullyParsedClass.interfaceRefs[i].classIndex),
				NameAndType: uint16(fullyParsedClass.interfaceRefs[i].nameAndTypeIndex),
			}
			kd.CP.InterfaceRefs = append(kd.CP.InterfaceRefs, ir)
		}
	}

	if len(fullyParsedClass.invokeDynamics) > 0 {
		for i := 0; i < len(fullyParsedClass.invokeDynamics); i++ {
			id := InvokeDynamicEntry{
				BootstrapIndex: uint16(fullyParsedClass.invokeDynamics[i].bootstrapIndex),
				NameAndType:    uint16(fullyParsedClass.invokeDynamics[i].nameAndType),
			}
			kd.CP.InvokeDynamics = append(kd.CP.InvokeDynamics, id)
		}
	}

	if len(fullyParsedClass.longConsts) > 0 {
		for i := 0; i < len(fullyParsedClass.longConsts); i++ {
			kd.CP.LongConsts = append(kd.CP.LongConsts, fullyParsedClass.longConsts[i])
		}
	}

	if len(fullyParsedClass.methodHandles) > 0 {
		for i := 0; i < len(fullyParsedClass.methodHandles); i++ {
			mh := MethodHandleEntry{
				RefKind:  uint16(fullyParsedClass.methodHandles[i].referenceKind),
				RefIndex: uint16(fullyParsedClass.methodHandles[i].referenceIndex),
			}
			kd.CP.MethodHandles = append(kd.CP.MethodHandles, mh)
		}
	}

	if len(fullyParsedClass.methodRefs) > 0 {
		for i := 0; i < len(fullyParsedClass.methodRefs); i++ {
			mr := MethodRefEntry{
				ClassIndex:  uint16(fullyParsedClass.methodRefs[i].classIndex),
				NameAndType: uint16(fullyParsedClass.methodRefs[i].nameAndTypeIndex),
			}
			kd.CP.MethodRefs = append(kd.CP.MethodRefs, mr)
		}
	}

	if len(fullyParsedClass.methodTypes) > 0 {
		for i := 0; i < len(fullyParsedClass.methodTypes); i++ {
			kd.CP.MethodTypes = append(kd.CP.MethodTypes, uint16(fullyParsedClass.methodTypes[i]))
		}
	}

	if len(fullyParsedClass.nameAndTypes) > 0 {
		for i := 0; i < len(fullyParsedClass.nameAndTypes); i++ {
			nat := NameAndTypeEntry{
				NameIndex: uint16(fullyParsedClass.nameAndTypes[i].nameIndex),
				DescIndex: uint16(fullyParsedClass.nameAndTypes[i].descriptorIndex),
			}
			kd.CP.NameAndTypes = append(kd.CP.NameAndTypes, nat)
		}
	}

	if len(fullyParsedClass.utf8Refs) > 0 {
		for i := 0; i < len(fullyParsedClass.utf8Refs); i++ {
			kd.CP.Utf8Refs = append(kd.CP.Utf8Refs, fullyParsedClass.utf8Refs[i].content)
		}
	}

	return kd
}

// GetCountOfLoadedClasses returns the number of classes loaded
// by the classloader
func (cl *Classloader) GetCountOfLoadedClasses() int {
	return cl.ClassCount
}

// accepts a string containing a class reference from a class file and converts
// it into a normalized z/y/x format. It converts references that start with [L
// and skips all array classes. For these latter cases or any errors, it returns ""
func normalizeClassReference(ref string) string {
	refClassName := ref
	if strings.HasPrefix(refClassName, types.RefArray) {
		refClassName = strings.TrimPrefix(refClassName, types.RefArray)
		if strings.HasSuffix(refClassName, ";") {
			refClassName = strings.TrimSuffix(refClassName, ";")
		}
	} else if strings.HasPrefix(refClassName, types.Array) {
		refClassName = ""
	}
	return refClassName
}

// Init simply initializes the three classloaders and the class area
// and points the classloaders to each other in the proper order.
// This function might be substantially revised later.
func Init() error {
	BootstrapCL.Name = "bootstrap"
	BootstrapCL.Parent = ""
	BootstrapCL.ClassCount = 0
	BootstrapCL.Archives = make(map[string]*Archive)

	ExtensionCL.Name = "extension"
	ExtensionCL.Parent = "bootstrap"
	ExtensionCL.ClassCount = 0
	ExtensionCL.Archives = make(map[string]*Archive)

	AppCL.Name = "app"
	AppCL.Parent = "extension"
	AppCL.ClassCount = 0
	AppCL.Archives = make(map[string]*Archive)

	// Launch JmodMap initialisation
	// commented out: go JmodMapInit()
	JmodMapInit()

	// Load the base jmod
	GetBaseJmodBytes()

	// initialize the method area
	InitMethodArea()

	// Success!
	return nil
}
