package pciinfo

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type Function struct {
	// TODO are these all the most interesting files?
	// need to choose best types (vs using all strings)
	Number string
	// SysFS Path:
	// /sys/devices/pciNNNN:NN/NNNN:NN:NN.N/
	//          .../pciDOMAIN:BUS/DOMAIN:BUS:DEVICE.FUNCTION
	Class           string `json:"class,omitempty"`            //		   PCI class (ascii, ro)
	Revision        string `json:"revision,omitempty"`         //PCI revision (ascii, ro)
	Vendor          string `json:"vendor,omitempty"`           //		   PCI vendor (ascii, ro)
	Device          string `json:"device,omitempty"`           //	   PCI device (ascii, ro)
	SubsystemVendor string `json:"subsystem_vendor,omitempty"` //	   PCI subsystem vendor (ascii, ro)
	SubsystemDevice string `json:"subsystem_device,omitempty"` //	   PCI subsystem device (ascii, ro)
	Enable          uint16 `json:"enable,omitempty"`           //big enough?           //	           Whether the device is enabled (ascii, rw)
	Irq             string `json:"irq,omitempty"`              //		   IRQ number (ascii, ro)
	LocalCpus       string `json:"local_cpus,omitempty"`       //	   nearby CPU mask (cpumask, ro)
	//Resource        string `json:"resource,omitempty"`         //		   PCI resource host addresses (ascii, ro)
	//Config     string //		   PCI config space (binary, rw)
	//Remove     string //		   remove device from kernel's list (ascii, wo)
	//Resource0..N	   PCI resource N, if present (binary, mmap, rw[1])
	//Resource0_wc..N_wc  PCI WC map resource N, if prefetchable (binary, mmap)
	//Rom              string //		   PCI ROM resource, if present (binary, ro)
	// the previous are the files described in the kernel.org docs:
	// https://www.kernel.org/doc/Documentation/filesystems/sysfs-pci.txt
	// but my computer has more files....
	//
	// TODO read 'uevent' instead of class/device/vendor files?
	// 'uevent' file has driver (in use), class, id (vendor:device
	// pair), subsystem id (v:d pair), pci slot, modalias.
	// Probably easier to pull from oneline files instead of chopping up
	// uevent lines
}

type Device struct {
	Number    string
	Functions map[string]*Function
}

type Bus struct {
	Number  string
	Devices map[string]*Device
}

type Domain struct {
	Number string
	Buses  map[string]*Bus
}

func getFunctionInfo(path string) (*Function, error) {
	fun := &Function{}
	infoFiles, err := ioutil.ReadDir(path)
	if err != nil {
		return fun, err
	}
	for _, f := range infoFiles {
		// At /sys/devices/pci0000:00/0000:00:00.0/*
		switch f.Name() {
		case "class":
			fun.Class = fileString(filepath.Join(path, f.Name()))
		case "vendor":
			fun.Vendor = fileString(filepath.Join(path, f.Name()))
		case "device":
			fun.Device = fileString(filepath.Join(path, f.Name()))
		case "enable":
			if enable, err := strconv.ParseUint(fileString(filepath.Join(path, f.Name())), 10, 16); err == nil {
				fun.Enable = uint16(enable)
			}
		//case "irq":
		//	fun.Irq = fileString(filepath.Join(path, f.Name()))
		//case "local_cpus":
		//	fun.LocalCpus = fileString(filepath.Join(path, f.Name()))
		//case "resource":
		//	fun.Resource = fileString(filepath.Join(path, f.Name()))
		case "revision":
			fun.Revision = fileString(filepath.Join(path, f.Name()))
		case "subsystem_vendor":
			fun.SubsystemVendor = fileString(filepath.Join(path, f.Name()))
		case "subsystem_device":
			fun.SubsystemDevice = fileString(filepath.Join(path, f.Name()))
		}
	}
	return fun, nil
}

func enumeratePci() (map[string]*Domain, error) {
	sysBase := "/sys/devices/"
	dirs, err := ioutil.ReadDir(sysBase)
	domains := map[string]*Domain{}
	if err != nil {
		return domains, err
	}
	for _, dir := range dirs {
		// At /sys/devices/pci0000:00/
		if strings.HasPrefix(dir.Name(), "pci") {
			domStr := strings.Split(dir.Name()[3:], ":")[0]
			domains[domStr] = &Domain{Number: domStr, Buses: map[string]*Bus{}}
			subDirs, err := ioutil.ReadDir(filepath.Join(sysBase, dir.Name()))
			if err != nil {
				return domains, err
			}

			for _, sDir := range subDirs {
				// At /sys/devices/pci0000:00/0000:00:00.0
				if strings.HasPrefix(sDir.Name(), domStr) {
					busEtcStrings := strings.Split(sDir.Name(), ":")[1:]
					bus, ok := domains[domStr].Buses[busEtcStrings[0]]
					if !ok {
						bus = &Bus{Number: busEtcStrings[0], Devices: map[string]*Device{}}
						domains[domStr].Buses[busEtcStrings[0]] = bus
					}
					devFuncStrings := strings.Split(busEtcStrings[1], ".")
					dev, ok := bus.Devices[devFuncStrings[0]]
					if !ok {
						dev = &Device{Number: devFuncStrings[0], Functions: map[string]*Function{}}
						bus.Devices[devFuncStrings[0]] = dev
					}
					function, err := getFunctionInfo(filepath.Join(sysBase, dir.Name(), sDir.Name()))
					if err != nil {
						return domains, err
					}
					function.Number = devFuncStrings[1]
					dev.Functions[function.Number] = function
				}
			}
		}
	}
	return domains, err
}

func Dump() {
	// debug: print all the maps
	// not printing very pretty
	domains, err := enumeratePci()
	if err != nil {
		fmt.Println(err)
	}
	for _, dom := range domains {
		for _, bus := range dom.Buses {
			for _, device := range bus.Devices {
				for _, fun := range device.Functions {
					fmt.Printf("Dom: %v Bus %v: Device %v: Func %v:", dom.Number, bus.Number, device.Number, fun.Number)
					fmt.Println(fun)
				}
			}
		}
	}
}

func DumpHuman() {
	// use pcidb to get human readable vendor/device strings, etc
}

func getVendorString(hexString string) string {
	// Function.Vendor --> human readable
	return ""
}
