package store

import (
	"encoding/json"

	engine "github.com/jimi36/app-engine"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	appkeyPath    = "/store/app"
	runtimePath   = "/store/runtime"
	configkeyPath = "/store/config"
)

func makeAppkey(tag string) []byte {
	key := appkeyPath + "/" + tag
	return []byte(key)
}

func makeRuntimekey(name string) []byte {
	key := runtimePath + "/" + name
	return []byte(key)
}

func makeConfigkey(name string) []byte {
	key := configkeyPath + "/" + name
	return []byte(key)
}

type LevelDBStore struct {
	db *leveldb.DB
}

var _ engine.Store = (*LevelDBStore)(nil)

func NewLevelDBStore(filePath string) (*LevelDBStore, error) {
	db, err := leveldb.OpenFile(filePath, nil)
	if err != nil {
		return nil, err
	}

	return &LevelDBStore{
		db: db,
	}, nil
}

func (s *LevelDBStore) AddApplication(app *engine.Application) error {
	data, err := json.Marshal(app)
	if err != nil {
		return err
	}

	key := makeAppkey(app.Tag())
	if has, _ := s.HasApplication(&app.ApplicationTag); has {
		return engine.ErrStoreAppExisted
	}
	if err := s.db.Put(key, data, nil); err != nil {
		return err
	}

	return nil
}

func (s *LevelDBStore) RemoveApplication(tag *engine.ApplicationTag) error {
	if has, _ := s.HasApplication(tag); !has {
		return engine.ErrStoreAppNoFound
	}
	if err := s.db.Delete(makeAppkey(tag.Tag()), nil); err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStore) UpdateApplication(tag *engine.ApplicationTag, f func(*engine.Application)) error {
	app, err := s.GetApplication(tag)
	if err != nil {
		return err
	}

	f(app)

	data, err := json.Marshal(app)
	if err != nil {
		return err
	}

	if err := s.db.Put(makeAppkey(tag.Tag()), data, nil); err != nil {
		return err
	}

	return nil
}

func (s *LevelDBStore) HasApplication(tag *engine.ApplicationTag) (bool, error) {
	has, err := s.db.Has(makeAppkey(tag.Tag()), nil)
	if err != nil {
		return false, err
	}
	return has, nil
}

func (s *LevelDBStore) GetApplication(tag *engine.ApplicationTag) (*engine.Application, error) {
	data, err := s.db.Get(makeAppkey(tag.Tag()), nil)
	if err != nil {
		return nil, err
	}

	app := &engine.Application{}
	if err := json.Unmarshal(data, app); err != nil {
		return nil, err
	}

	return app, nil
}

func (s *LevelDBStore) ListApplications(size int, lastPos string) ([]*engine.ApplicationTag, string, error) {
	newPos := lastPos
	var out []*engine.ApplicationTag
	iter := s.db.NewIterator(util.BytesPrefix([]byte(appkeyPath)), nil)
	iter.Seek([]byte(lastPos))
	for iter.Next() && size > 0 {
		app := &engine.Application{}
		if err := json.Unmarshal(iter.Value(), &app); err != nil {
			return nil, "", err
		}

		out = append(out, &engine.ApplicationTag{app.Name, app.Version})
		newPos = string(makeAppkey(app.Name))

		size--
	}
	iter.Release()

	return out, newPos, nil
}

func (s *LevelDBStore) AddApplicationRunTime(rt *engine.ApplicationRuntime) error {
	data, err := json.Marshal(rt)
	if err != nil {
		return err
	}

	if err := s.db.Put(makeRuntimekey(rt.Name), data, nil); err != nil {
		return err
	}

	return nil
}

func (s *LevelDBStore) RemoveApplicationRunTime(name string) error {
	if err := s.db.Delete(makeRuntimekey(name), &opt.WriteOptions{Sync: true}); err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStore) UpdateApplicationRuntime(name string, f func(*engine.ApplicationRuntime) error) error {
	data, err := s.db.Get(makeRuntimekey(name), nil)
	if err != nil {
		return err
	}

	rt := &engine.ApplicationRuntime{}
	if err := json.Unmarshal(data, rt); err != nil {
		return err
	}

	if err := f(rt); err != nil {
		return err
	}

	data, err = json.Marshal(rt)
	if err != nil {
		return err
	}

	if err := s.db.Put(makeRuntimekey(name), data, nil); err != nil {
		return err
	}

	return nil
}

func (s *LevelDBStore) GetApplicationRuntime(name string) (*engine.ApplicationRuntime, error) {
	data, err := s.db.Get(makeRuntimekey(name), nil)
	if err != nil {
		return nil, engine.ErrStoreAppRuntimeNoFound
	}

	rt := &engine.ApplicationRuntime{}
	if err := json.Unmarshal(data, rt); err != nil {
		return nil, err
	}

	return rt, nil
}

func (s *LevelDBStore) ListApplicationRunTimes(size int, lastPos string) ([]*engine.ApplicationRuntime, string, error) {
	iter := s.db.NewIterator(util.BytesPrefix([]byte(runtimePath)), nil)
	if len(lastPos) > 0 {
		iter.Seek([]byte(lastPos))
	}

	newPos := lastPos
	var out []*engine.ApplicationRuntime
	for iter.Next() && size > 0 {
		rt := &engine.ApplicationRuntime{}
		if err := json.Unmarshal(iter.Value(), rt); err != nil {
			continue
		}
		out = append(out, rt)
		newPos = string(makeRuntimekey(rt.Name))
	}
	iter.Release()

	return out, newPos, iter.Error()
}

func (s *LevelDBStore) ForeachApplicationRunTime(f func(*engine.ApplicationRuntime)) error {
	iter := s.db.NewIterator(util.BytesPrefix([]byte(runtimePath)), nil)
	for iter.Next() {
		rt := &engine.ApplicationRuntime{}
		if err := json.Unmarshal(iter.Value(), rt); err != nil {
			continue
		}
		f(rt)
	}
	iter.Release()

	return iter.Error()
}

func (s *LevelDBStore) AddConfig(config *engine.Config) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	if err := s.db.Put(makeConfigkey(config.Name), data, nil); err != nil {
		return err
	}

	return nil
}

func (s *LevelDBStore) RemoveConfig(name string) error {
	if err := s.db.Delete(makeConfigkey(name), nil); err != nil {
		return err
	}
	return nil
}

func (s *LevelDBStore) HasConfig(name string) (bool, error) {
	return s.db.Has(makeConfigkey(name), nil)
}

func (s *LevelDBStore) GetConfig(name string) (*engine.Config, error) {
	data, err := s.db.Get(makeConfigkey(name), nil)
	if err != nil {
		return nil, err
	}

	config := &engine.Config{}
	if err := json.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}
