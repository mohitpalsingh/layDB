package main

func (db *LayDB) load() error {
	if db.log == nil {
		return nil
	}
	i := 0
	for {
		data, err := db.log.Read(uint64(i))
		if err != nil {
			if err == ErrEOF {
				break
			}
			return err
		}

		record, err := decode(data)
		if err != nil {
			return err
		}

		if len(record.meta.key) > 0 {
			if err := db.loadRecord(record); err != nil {
				return err
			}
		}

		i++
	}
	return nil
}

func (db *LayDB) loadRecord(r *record) (err error) {
	err = db.buildStringRecord(r)
	return
}

func (db *LayDB) buildStringRecord(r *record) (err error) {
	key := string(r.meta.key)
	db.strStore.Insert([]byte(key), r.meta.value)

	return nil
}
