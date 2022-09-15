package gopythonutil

import (
	"fmt"
	"github.com/DataDog/go-python3"
	"os"
)

// 封装 go-python3 一些比较常用的方法
func Initialize() {
	python3.Py_Initialize()
	if !python3.Py_IsInitialized(){
		fmt.Println("Error initializing the python interpreter")
		os.Exit(1)
	}
}

func ImportSystemModule() {
	sysModule := python3.PyImport_ImportModule("sys")
	path := sysModule.GetAttrString("path")
	//python3.PyList_Insert(path, 0, python3.PyUnicode_FromString("/Library/Frameworks/Python.framework/Versions/3.7/lib/python3.7/site-packages"))
	python3.PyList_Insert(path, 0, python3.PyUnicode_FromString("/usr/local/lib/python3.7/dist-packages"))
	python3.PyList_Insert(path, 0, python3.PyUnicode_FromString("/usr/lib/python3.7"))
}

func ImportCustomModule(dir string) {
	sysModule := python3.PyImport_ImportModule("sys")
	path := sysModule.GetAttrString("path")
	python3.PyList_Insert(path, 0, python3.PyUnicode_FromString(dir))
}

func GetModule(name string) *python3.PyObject{
	return python3.PyImport_ImportModule(name)
}


func GetEmptyTuple() *python3.PyObject {
	return python3.PyTuple_New(0)
}
