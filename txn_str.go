package laydb

// sets the key value pair
func (tx *Tx) Set(key string, value string) error {
	e := newRecordWithValue([]byte(key), []byte(value))
	tx.addRecord(e)
	return nil
}

// returns the value for the given key
func (tx *Tx) Get(key string) (val string, err error) {
	val, err = tx.get(key)
	if err != nil {
		return
	}
	return
}

// deletes the given key
func (tx *Tx) Delete(key string) error {
	e := newRecordWithValue([]byte(key), nil)
	tx.addRecord(e)
	return nil
}

// checks whether a key exists in DB
func (tx *Tx) Exists(key string) bool {
	_, err := tx.db.strStore.get(key)
	if err != nil {
		return false
	}
	return true
}

// helper func to fetch value of a key from DB
func (tx *Tx) get(key string) (val string, err error) {
	v, err := tx.db.strStore.get(key)
	if err != nil {
		return "", err
	}
	val = v.(string)
	return
}
