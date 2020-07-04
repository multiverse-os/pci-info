package pciinfo

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type Function struct {
	Number string
	// SysFS Path:
	// /sys/devices/pciNNNN:NN/NNNN:NN:NN.N/
	//          .../pciDOMAIN:BUS/DOMAIN:BUS:DEVICE.FUNCTION
	Class           string `json:"class,omitempty"`            //		   PCI class (ascii, ro)
	Device          string `json:"device,omitempty"`           //	   PCI device (ascii, ro)
	Enable          uint16 `json:"enable,omitempty"`           //big enough?           //	           Whether the device is enabled (ascii, rw)
	Irq             string `json:"irq,omitempty"`              //		   IRQ number (ascii, ro)
	LocalCpus       string `json:"local_cpus,omitempty"`       //	   nearby CPU mask (cpumask, ro)
	Resource        string `json:"resource,omitempty"`         //		   PCI resource host addresses (ascii, ro)
	Revision        string `json:"revision,omitempty"`         //PCI revision (ascii, ro)
	SubsystemDevice string `json:"subsystem_device,omitempty"` //	   PCI subsystem device (ascii, ro)
	SubsystemVendor string `json:"subsystem_vendor,omitempty"` //	   PCI subsystem vendor (ascii, ro)
	Vendor          string `json:"vendor,omitempty"`           //		   PCI vendor (ascii, ro)
	//Config     string //		   PCI config space (binary, rw)
	//Remove     string //		   remove device from kernel's list (ascii, wo)
	//Resource0..N	   PCI resource N, if present (binary, mmap, rw[1])
	//Resource0_wc..N_wc  PCI WC map resource N, if prefetchable (binary, mmap)
	//Rom              string //		   PCI ROM resource, if present (binary, ro)
	// the previous are the files described in the kernel.org docs:
	// https://www.kernel.org/doc/Documentation/filesystems/sysfs-pci.txt
	// but my computer has more files....
}

type Device struct {
	Number    string
	Functions map[string]Function
}

type Bus struct {
	Number  string
	Devices map[string]Device
}

type Domain struct {
	Number string
	Buses  map[string]Bus
}

func getFunctionInfo(dom, bus, dev, fun string) Function {
	return Function{}
}

func enumeratePci() (map[string]Domain, error) {
	// getPciInfo for every pci device
	sysBase := "/sys/devices/"
	dirs, err := ioutil.ReadDir(sysBase)
	domains := map[string]Domain{}
	if err != nil {
		return domains, err
	}
	for _, dir := range dirs {
		if strings.HasPrefix(dir.Name(), "pci") {
			domStr := strings.Split(dir.Name()[3:], ":")[0]
			domains[domStr] = Domain{Number: domStr}
			subDirs, err := ioutil.ReadDir(filepath.Join(sysBase, dir.Name()))
			if err != nil {
				return domains, err
			}

			for _, sDir := range subDirs {
				if strings.HasPrefix(sDir.Name(), domStr) {
					infoFiles, err := ioutil.ReadDir(filepath.Join(sysBase, dir.Name(), sDir.Name()))
					if err != nil {
						return domains, err
					}
					bdfStrings := strings.Split(sDir.Name(), ":")[1:]
					devfunc := strings.Split(bdfStrings[1], ".")
					bdfStrings = []string{bdfStrings[0], devfunc[0], devfunc[1]}
					function := Function{Number: bdfStrings[2]}
					for _, f := range infoFiles {
						switch f.Name() {
						case "class":
							function.Class = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "device":
							function.Device = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "enable":
							if enable, err := strconv.ParseUint(fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name())), 10, 16); err == nil {
								function.Enable = uint16(enable)
							}
						case "irq":
							function.Irq = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "local_cpus":
							function.LocalCpus = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "resource":
							function.Resource = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "revision":
							function.Revision = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "subsystem_device":
							function.SubsystemDevice = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "subsystem_vendor":
							function.SubsystemVendor = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						case "vendor":
							function.Vendor = fileString(filepath.Join(sysBase, dir.Name(), sDir.Name(), f.Name()))
						}
					}
					if bus, ok := domains[domStr].Buses[bdfStrings[0]]; ok {
						fmt.Println("if 1")
						if device, ok := bus.Devices[bdfStrings[1]]; ok {
							fmt.Println("if 2")
							if funct, ok := device.Functions[function.Number]; ok {
								fmt.Println("if 3")
								fmt.Printf("redundant function:\n%v\n", funct)
								//this shouldn't happen, functions shouldn't be repeated in a
								//device
							} else {
								fmt.Println("else 3")
								funct = function
							}
						} else {
							fmt.Println("else 2")
							funcs := map[string]Function{function.Number: function}
							device = Device{Number: bdfStrings[1], Functions: funcs}
							devices := map[string]Device{device.Number: device}
							//domains[domStr].Buses[bdfStrings[0]].Devices = map[string]Device{device.Number: device} // TODO breaks
							// cannot assign to struct field domains[domStr].Buses[bdfStrings[0]].Devices in map
							// there has to be a better way!
						}
					} else {
						fmt.Println("else 1")
						funcs := map[string]Function{function.Number: function}
						device := Device{Number: bdfStrings[1], Functions: funcs}
						devices := map[string]Device{device.Number: device}
						bus = Bus{Number: bdfStrings[0], Devices: devices}
						dom := domains[domStr]
						dom.Buses = map[string]Bus{bus.Number: bus}
						domains[domStr] = dom
					}

				}
			}
		}
	}
	fmt.Println(domains["0000"].Buses)
	return domains, err
}

func Dump() {
	enumeratePci()
}

func DumpHuman() {
	// use pcidb to get human readable vendor/device strings, etc
}

func getVendorString(hexString string) string {
	return ""
}
