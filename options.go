package loggable

import (
	"reflect"
)

// Option is a generic options pattern.
type Option func(options *options)

type options struct {
	lazyUpdate       bool
	lazyUpdateFields []string
	metaTypes        map[string]reflect.Type
	objectTypes      map[string]reflect.Type
	computeDiff      bool
}

// Option ComputeDiff allows you also write differences between objects on update operations.
// ComputeDiff not reads records from db, it used only as cache on plugin side.
// So it does not track changes outside plugin.
func ComputeDiff() Option {
	return func(options *options) {
		options.computeDiff = true
	}
}

// Option LazyUpdate allows you to skip update operations when nothing was changed.
// Parameter 'fields' is list of sql field names that should be ignored on updates.
func LazyUpdate(fields ...string) Option {
	return func(options *options) {
		options.lazyUpdate = true
		options.lazyUpdateFields = fields
	}
}

// RegObjectType maps object to type name, that is used in field Type of ChangeLog struct.
// On read change log operations, if plugin finds registered object type, by its name from db,
// it unmarshal field RawObject to Object field via json.Unmarshal.
//
// To access decoded object, e.g. `ReallyFunnyClient`, use type casting: `changeLog.Object.(ReallyFunnyClient)`.
func RegObjectType(objectType string, objectStruct interface{}) Option {
	return func(options *options) {
		if options.objectTypes == nil {
			options.objectTypes = make(map[string]reflect.Type)
		}
		options.objectTypes[objectType] = reflect.Indirect(reflect.ValueOf(objectStruct)).Type()
	}
}

func RegObjectTypes(objectMap map[string]interface{}) Option {
	return func(options *options) {
		if options.objectTypes == nil {
			options.objectTypes = make(map[string]reflect.Type)
		}

		for t, v := range objectMap {
			options.objectTypes[t] = reflect.Indirect(reflect.ValueOf(v)).Type()
		}
	}
}

// RegMetaType works like RegObjectType, but for field RawMeta.
// RegMetaType maps object to type name, that is used in field Type of ChangeLog struct.
// On read change log operations, if plugin finds registered object type, by its name from db,
// it unmarshal field RawMeta to Meta field via json.Unmarshal.
//
// To access decoded object, e.g. `MyClientMeta`, use type casting: `changeLog.Meta.(MyClientMeta)`.
func RegMetaType(objectType string, metaType interface{}) Option {
	return func(options *options) {
		if options.metaTypes == nil {
			options.metaTypes = make(map[string]reflect.Type)
		}
		options.metaTypes[objectType] = reflect.Indirect(reflect.ValueOf(metaType)).Type()
	}
}
