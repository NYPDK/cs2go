package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"reflect"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func getProcessHandle(pid int) (windows.Handle, error) {
	// windows.PROCESS_VM_READ|windows.PROCESS_VM_WRITE|windows.PROCESS_VM_OPERATION
	return windows.OpenProcess(windows.PROCESS_VM_READ, false, uint32(pid))
}

func findProcessId(name string) (int, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0, err
	}

	defer windows.CloseHandle(windows.Handle(snapshot))

	for {
		var process windows.ProcessEntry32
		process.Size = uint32(unsafe.Sizeof(process))
		if windows.Process32Next(windows.Handle(snapshot), &process) != nil {
			break
		}
		if windows.UTF16ToString(process.ExeFile[:]) == name {
			return int(process.ProcessID), nil
		}
	}
	return 0, fmt.Errorf("module not found")
}

func getModuleBaseAddress(pid int, moduleName string) (uintptr, error) {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPMODULE, uint32(pid))
	if err != nil {
		return 0, err
	}
	defer windows.CloseHandle(snapshot)

	var me32 windows.ModuleEntry32
	me32.Size = uint32(unsafe.Sizeof(me32))

	if windows.Module32First(snapshot, &me32) != nil {
		return 0, fmt.Errorf("Module32First failed")
	}

	for {
		if strings.EqualFold(windows.UTF16ToString(me32.Module[:]), moduleName) {
			return uintptr(me32.ModBaseAddr), nil
		}
		if windows.Module32Next(snapshot, &me32) != nil {
			break
		}
	}

	return 0, fmt.Errorf("module not found")
}

func read(process windows.Handle, address uintptr, value interface{}) error {
	var buffer []byte
	size := int(reflect.TypeOf(value).Elem().Size())
	buffer = make([]byte, size)
	bytesRead := uintptr(0)
	err := windows.ReadProcessMemory(process, address, &buffer[0], uintptr(len(buffer)), &bytesRead)
	if err != nil {
		return err
	}
	if bytesRead != uintptr(len(buffer)) {
		return fmt.Errorf("read %d bytes, expected %d", bytesRead, len(buffer))
	}
	switch v := value.(type) {
	case *int32:
		*v = int32(binary.LittleEndian.Uint32(buffer))
	case *uint32:
		*v = binary.LittleEndian.Uint32(buffer)
	case *float32:
		*v = math.Float32frombits(binary.LittleEndian.Uint32(buffer))
	case *int64:
		*v = int64(binary.LittleEndian.Uint64(buffer))
	case *uint64:
		*v = binary.LittleEndian.Uint64(buffer)
	case *float64:
		*v = math.Float64frombits(binary.LittleEndian.Uint64(buffer))
	case *uintptr:
		*v = uintptr(binary.LittleEndian.Uint64(buffer))
	case *Vector3:
		v.X = math.Float32frombits(binary.LittleEndian.Uint32(buffer[0:4]))
		v.Y = math.Float32frombits(binary.LittleEndian.Uint32(buffer[4:8]))
		v.Z = math.Float32frombits(binary.LittleEndian.Uint32(buffer[8:12]))
	default:
		err = binary.Read(bytes.NewReader(buffer), binary.LittleEndian, value)
		if err != nil {
			return err
		}
	}
	return nil
}

// func write(process windows.Handle, address uintptr, value interface{}) error {
// 	var buffer bytes.Buffer
// 	err := binary.Write(&buffer, binary.LittleEndian, value)
// 	if err != nil {
// 		return err
// 	}
// 	bytesWritten := uintptr(0)
// 	err = windows.WriteProcessMemory(process, address, &buffer.Bytes()[0], uintptr(buffer.Len()), &bytesWritten)
// 	if err != nil {
// 		return err
// 	}
// 	if bytesWritten != uintptr(buffer.Len()) {
// 		return fmt.Errorf("wrote %d bytes, expected %d", bytesWritten, buffer.Len())
// 	}
// 	return nil
// }
