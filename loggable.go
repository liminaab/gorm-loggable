package loggable

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/gofrs/uuid"
	"github.com/jinzhu/gorm"
)

// Interface is used to get metadata from your models.
type Interface interface {
	// Meta should return structure, that can be converted to json.
	Meta() interface{}
	// lock makes available only embedding structures.
	lock()
	// check if callback enabled
	isEnabled() bool
	// enable/disable loggable
	Enable(v bool)
	//
	SecondaryIndexValue() string
}

// LoggableModel is a root structure, which implement Interface.
// Embed LoggableModel to your model so that Plugin starts tracking changes.
type LoggableModel struct {
	Disabled bool `sql:"-" json:"-"`
	SecondaryIndex string `sql:"-" json:"-"`
}

func (LoggableModel) Meta() interface{} { return nil }
func (LoggableModel) lock()             {}
func (l LoggableModel) isEnabled() bool { return !l.Disabled }
func (l *LoggableModel) Enable(v bool)   { l.Disabled = !v }
func (l LoggableModel) SecondaryIndexValue() string { return "" }

type Changelogs []ChangeLog

// ChangeLog is a main entity, which used to log changes.
// Commonly, ChangeLog is stored in 'change_logs' table.
type ChangeLog struct {
	// Primary key of change logs.
	ID uuid.UUID `gorm:"primary_key;" json:"id"`
	// Timestamp, when change log was created.
	CreatedAt time.Time `sql:"DEFAULT:current_timestamp" json:"created_at"`
	// Action type.
	// On write, supports only 'create', 'update', 'delete',
	// but on read can be anything.
	Action string `json:"action"`
	// ID of tracking object.
	// By this ID later you can find all object (database row) changes.
	ObjectID string `gorm:"index" json:"object_id"`
	ObjectID2 string `gorm:"index" json:"object_id2"`
	// Reflect name of tracking object.
	// It does not use package or module name, so
	// it may be not unique when use multiple types from different packages but with the same name.
	ObjectType string `gorm:"index" json:"object_type"`
	// Raw representation of tracking object.
	// todo(@sas1024): Replace with []byte, to reduce allocations. Would be major version.
	RawObject string `sql:"type:JSON" json:"raw_object"`
	// Raw representation of tracking object's meta.
	// todo(@sas1024): Replace with []byte, to reduce allocations. Would be major version.
	RawMeta string `sql:"type:JSON" json:"raw_meta"`
	// Raw representation of diff's.
	// todo(@sas1024): Replace with []byte, to reduce allocations. Would be major version.
	RawDiff string `sql:"type:JSON" json:"raw_diff"`
	// Free field to store something you want, e.g. who creates change log.
	// Not used field in gorm-loggable, but gorm tracks this field.
	CreatedBy string `gorm:"index" json:"created_by"`
	// Field Object would contain prepared structure, parsed from RawObject as json.
	// Use RegObjectType to register object types.
	Object interface{} `sql:"-" json:"object"`
	// Field Meta would contain prepared structure, parsed from RawMeta as json.
	// Use RegMetaType to register object's meta types.
	Meta interface{} `sql:"-" json:"meta"`
}

func (l *ChangeLog) prepareObject(objType reflect.Type) error {
	// Allocate new and try to decode change logs field RawObject to Object.
	obj := reflect.New(objType).Interface()
	err := json.Unmarshal([]byte(l.RawObject), obj)
	l.Object = obj
	return err
}

func (l *ChangeLog) prepareMeta(objType reflect.Type) error {
	// Allocate new and try to decode change logs field RawObject to Object.
	obj := reflect.New(objType).Interface()
	err := json.Unmarshal([]byte(l.RawMeta), obj)
	l.Meta = obj
	return err
}

// Diff returns parsed to map[string]interface{} diff representation from field RawDiff.
// To unmarshal diff to own structure, manually use field RawDiff.
func (l ChangeLog) Diff() (UpdateDiff, error) {
	var diff UpdateDiff
	err := json.Unmarshal([]byte(l.RawDiff), &diff)
	if err != nil {
		return nil, err
	}

	return diff, nil
}

func interfaceToString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case *string:
		if val != nil {
			return *val
		} else {
			return ""
		}
	case int64:
		return fmt.Sprint(val)
	case *int64:
		if val != nil {
			return fmt.Sprint(*val)
		} else {
			return ""
		}
	default:
		return fmt.Sprint(v)
	}
}

func fetchChangeLogMeta(scope *gorm.Scope) []byte {
	val, ok := scope.Value.(Interface)
	if !ok {
		return nil
	}
	data, err := json.Marshal(val.Meta())
	if err != nil {
		panic(err)
	}
	return data
}

func isLoggable(value interface{}) bool {
	_, ok := value.(Interface)
	return ok
}

func isEnabled(value interface{}) bool {
	v, ok := value.(Interface)
	return ok && v.isEnabled()
}
