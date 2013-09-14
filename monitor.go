/*
Package fsUtils contains a set of useful utilities I have developed for working with the filesystem from go programs. Its code base will grow on an "as needed" basis.
*/
package fsUtils

import (
	"io/ioutil"
	"time"
)

/*
Monitor is a structure that keeps track of the contents of a directory alerting the program when changes occure.

An example of how to monitor a directory named "test":
	func main() {
		var m fsUtils.Monitor
		err := m.Directory("test",testAdd,testDel)
		if err!=nil {
			fmt.Println(err.Error())
		}
	}

	func testAdd(file string) {
		fmt.Println("Added "+file)
	}

	func testDel(file string) {
		fmt.Println("Deleted "+file)
	}
*/
type Monitor struct {
	contents map[string]bool
}

type change struct {
	Name    string
	Deleted bool
}

/*
Directory causes a Monitor to begin monitoring a directory, calling the onAdd and onDelete callback functions when a change is detected.
*/
func (m *Monitor) Directory(directoryName string, onAdd func(string), onDelete func(string)) error {
	err := m.buildContents(directoryName)
	if err != nil {
		return err
	}
	handlechanges(m.contentArray(),onAdd,nil)

	for {
		time.Sleep(1000 * time.Millisecond)
		change, err := m.getDiff(directoryName)
		if err != nil {
			return err
		}
		if len(change) > 0 {
			handlechanges(change,onAdd,onDelete)
		}
	}

	return nil
}

func handlechanges(changes []change, onAdd func(string), onDelete func(string)) {
	for _,change := range changes {
		if change.Deleted {
			onDelete(change.Name)
		} else {
			onAdd(change.Name)
		}
	}
}

func (m *Monitor) buildContents(directoryName string) error {
	folder, err := ioutil.ReadDir(directoryName)

	if err != nil {
		return err
	}

	m.contents = make(map[string]bool)
	for _, file := range folder {
		m.contents[file.Name()] = false
	}
	return nil
}

func (m *Monitor) contentArray() []change {
	result := make([]change, len(m.contents))
	i := 0
	for key, _ := range m.contents {
		result[i] = change{key, false}
		i++
	}
	return result
}

func (m *Monitor) getDiff(directoryName string) ([]change, error) {
	folder, err := ioutil.ReadDir(directoryName)
	result := make([]change, 0, len(folder)+len(m.contents))

	if err != nil {
		return nil, err
	}

	i := 0 //index for result

	//Ensure files are in contents already
	for _, file := range folder {
		_, ok := m.contents[file.Name()]
		if !ok {
			m.contents[file.Name()] = true
			result = result[0 : len(result)+1]
			result[i] = change{file.Name(), false}
			i++
		} else {
			m.contents[file.Name()] = true
		}
	}

	//Check if files have been removed
	for key, value := range m.contents {
		if !value {
			delete(m.contents, key)
			result = result[0 : len(result)+1]
			result[i] = change{key, true}
			i++
		}
	}

	//Set files back to false
	for key, _ := range m.contents {
		m.contents[key] = false
	}

	return result, nil
}
