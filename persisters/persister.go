package persisters

type ID int

type Persister interface {
	Create(object interface{}) (ID, error)
	Retrieve(id ID, template interface{}) error
	Update(id ID, object interface{}) error
	Delete(id ID) error
	List() map[ID]interface{}
}
