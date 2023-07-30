package persisters

import (
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
)

var dataPath = "%s/%v.json"
var filePermission = fs.FileMode(0644)

// map[Object Type: referenced by directory name]
// => headID (for new ID generation)
var objectHeadID map[string]ID

// map[Object Type: referenced by directory name]
// => object listing[active/inactive] (for fast listings)
var objectListing map[string]map[ID]interface{}

type FilePersister struct {
	fileDirectory string
}

func MakeFilePersister(objectType string) (FilePersister, error) {
	var fp FilePersister
	fp.fileDirectory = objectType

	objectHeadID = make(map[string]ID)
	if objectListing == nil {
		objectListing = make(map[string]map[ID]interface{})
	}
	objectListing[fp.fileDirectory] = make(map[ID]interface{})

	fp.scanDirectoryForIndex(objectType, objectListing)

	directoryExists, err := fileExists(fp.fileDirectory)
	if err != nil {
		return fp, err
	}
	if !directoryExists {
		err = os.MkdirAll(fp.fileDirectory, os.ModePerm)
		if err != nil {
			return fp, err
		}
	}

	return fp, nil
}

func (fp *FilePersister) Create(object interface{}) (ID, error) {
	id := fp.getNewID()

	fileName := fmt.Sprintf(dataPath, fp.fileDirectory, id)
	file, err := json.MarshalIndent(object, "", "")
	if err != nil {
		return id, err
	}
	err = os.WriteFile(fileName, file, filePermission)
	if err != nil {
		return id, err
	}

	objectListing[fp.fileDirectory][ID(id)] = object

	return id, nil
}

func (fp *FilePersister) Retrieve(id ID, objectTemplate interface{}) error {
	fileName := fmt.Sprintf(dataPath, fp.fileDirectory, id)
	jsonFile, err := os.Open(fileName)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	fileBytes, err := io.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	json.Unmarshal(fileBytes, &objectTemplate)

	return nil
}

func (fp *FilePersister) Update(id ID, object interface{}) error {
	idStr := strconv.Itoa(int(id))
	fileName := fmt.Sprintf(dataPath, fp.fileDirectory, idStr)
	file, err := json.MarshalIndent(object, "", "")
	if err != nil {
		return err
	}
	err = os.WriteFile(fileName, file, filePermission)
	if err != nil {
		return err
	}

	return nil
}

func (fp *FilePersister) Delete(id ID) error {
	idStr := strconv.Itoa(int(id))
	fileName := fmt.Sprintf(dataPath, fp.fileDirectory, idStr)
	err := os.Remove(fileName)
	if err != nil {
		return err
	}

	delete(objectListing[fp.fileDirectory], id)

	return nil
}

func (fp *FilePersister) List() map[ID]interface{} {
	return objectListing[fp.fileDirectory]
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (fp *FilePersister) getNewID() ID {
	ID, exists := objectHeadID[fp.fileDirectory]
	if !exists {
		objectHeadID[fp.fileDirectory] = 0
	}
	if ID == 0 {
		objectHeadID[fp.fileDirectory] = fp.searchForHighestID()
	}

	objectHeadID[fp.fileDirectory] = objectHeadID[fp.fileDirectory] + 1

	return objectHeadID[fp.fileDirectory]
}

func (fp *FilePersister) searchForHighestID() ID {
	files, err := ioutil.ReadDir(fp.fileDirectory)
	if err != nil {
		panic(err)
	}

	var maxID ID
	for _, file := range files {
		fileID := strings.Replace(file.Name(), ".json", "", 1)
		currentID, err := strconv.Atoi(fileID)
		if err != nil {
			panic(err)
		}
		if currentID > int(maxID) {
			maxID = ID(currentID)
		}
	}

	return maxID
}

func (fp *FilePersister) scanDirectoryForIndex(dirName string, index map[string]map[ID]interface{}) {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		return // directory not created yet
	}

	for _, file := range files {
		id := strings.ReplaceAll(file.Name(), ".json", "")
		idInt, err := strconv.Atoi(id)
		if err != nil {
			log.Fatal(err)
		}
		repo, exists := index[fp.fileDirectory]
		if !exists {
			index[fp.fileDirectory] = make(map[ID]interface{})
		}

		// Index can't provide more helpful data atm :/
		repo[ID(idInt)] = true
	}
}
