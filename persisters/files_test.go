package persisters

import (
	"os"
	"testing"
)

type person struct {
	Name  string   `json:"name"`
	Age   int      `json:"age"`
	Likes []string `json:"likes"`
}

func getTestObject1() person {
	return person{
		Name: "Ada Lovegood",
		Age:  18,
		Likes: []string{
			"math", "science", "computers",
			"programming", "sex"},
	}
}

func getTestObject2() person {
	return person{
		Name: "Issac Newton",
		Age:  52,
		Likes: []string{
			"math", "science", "physics"},
	}
}

func makeFilePersister(t *testing.T) FilePersister {
	fp, err := MakeFilePersister("tests")
	if err != nil {
		t.Fatal(err)
	}
	return fp
}

func TestFilePersisterReadWrite(t *testing.T) {
	fp := makeFilePersister(t)
	testPerson := getTestObject1()
	id, err := fp.Create(testPerson)
	if err != nil {
		t.Fatal(err)
	}
	var personFetch person
	err = fp.Retrieve(id, &personFetch)
	if err != nil {
		t.Fatal(err)
	}
	if personFetch.Age != testPerson.Age {
		m := "Create(person.age = 18), Sent: %v, Got %v"
		t.Fatalf(m, testPerson, personFetch)
	}
}

func TestFilePersisterUpdate(t *testing.T) {
	fp := makeFilePersister(t)

	var originalPerson person
	err := fp.Retrieve(1, &originalPerson)
	if err != nil {
		t.Fatal(err)
	}

	originalPerson.Age = originalPerson.Age + 1
	err = fp.Update(1, originalPerson)
	if err != nil {
		t.Fatal(err)
	}

	var updatedPerson person
	err = fp.Retrieve(1, &updatedPerson)
	if err != nil {
		t.Fatal(err)
	}

	if originalPerson.Age != updatedPerson.Age {
		m := "Update(person.age = 19), Local Sent: %v, Remote Value Post Update: %v"
		t.Fatalf(m, originalPerson.Age, updatedPerson.Age)
	}
}

func TestFilePersisterDelete(t *testing.T) {
	fp := makeFilePersister(t)

	err := fp.Delete(1)
	if err != nil {
		t.Fatal(err)
	}

	fileName := "tests/" + fp.fileDirectory + "/1.json"
	exists, err := fileExists(fileName)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Fatal("Object/File still exists: " + fileName)
	}
}

func TestManyObjects(t *testing.T) {
	fp := makeFilePersister(t)
	id1, err := fp.Create(getTestObject1())
	if err != nil {
		t.Fatal(err)
	}
	id2, err := fp.Create(getTestObject2())
	if err != nil {
		t.Fatal(err)
	}
	if id1 >= id2 {
		t.Fatal("ID did not increment")
	}
}

func TestObjectListing(t *testing.T) {
	fp := makeFilePersister(t)
	listing := fp.List()

	_, one := listing[1]
	_, two := listing[2]

	if !(one && two) {
		t.Fatal("Missing listings ^")
	}

	cleanTests(fp, t)
}

func cleanTests(fp FilePersister, T *testing.T) {
	err := os.RemoveAll("tests")
	if err != nil {
		T.Fatal(err)
	}
}
