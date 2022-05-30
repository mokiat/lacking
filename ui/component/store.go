package component

import (
	"fmt"
	"reflect"

	"github.com/mokiat/lacking/ui"
)

var (
	stores         map[reflect.Type]*Store
	connectedNodes []*componentNode
)

func init() {
	stores = make(map[reflect.Type]*Store)
}

// CreateStore creates a new Store instance.
//
// The specified reducer will be responsible for handling dispatched
// actions and adjusting the store's state.
//
// The initialValue should be of the same type that the reducer will operate
// on and can be used to initialize the Store.
//
// Note that similarly to Contexts, there can only be one Store instance
// per value type. Unlike Contexts, however, Stores are intended to have
// their value change throughout the lifecycle of the application.
func CreateStore(reducer Reducer, initialValue interface{}) *Store {
	result := &Store{
		reducer: reducer,
		value:   initialValue,
	}
	if _, ok := stores[reflect.TypeOf(initialValue)]; ok {
		panic(fmt.Errorf("there is already a store for values of type %T", initialValue))
	}
	stores[reflect.TypeOf(initialValue)] = result
	return result
}

// Store represents a state that can cross component boundaries, which allows
// multiple components to share the same source of truth.
type Store struct {
	reducer Reducer
	value   interface{}
}

// Get returns the value stored in the Store.
func (s *Store) Get() interface{} {
	return s.value
}

// Inject is a helper function that allows one to inject the Store's value
// directly into a variable referenced via the target pointer.
func (s *Store) Inject(target interface{}) {
	if target == nil {
		panic(fmt.Errorf("target cannot be nil"))
	}
	value := reflect.ValueOf(target)
	valueType := value.Type()
	if valueType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("target %T must be a pointer", target))
	}
	if value.IsNil() {
		panic(fmt.Errorf("target pointer cannot be nil"))
	}
	stateType := reflect.TypeOf(s.value)
	if !stateType.AssignableTo(valueType.Elem()) {
		panic(fmt.Errorf("cannot store value %T to specified type %s", s.value, valueType.Elem()))
	}
	value.Elem().Set(reflect.ValueOf(s.value))
}

// Destroy releases this Store and the value that it manages. A destroyed
// store will no longer be discoverable.
func (s *Store) Destroy() {
	if _, ok := stores[reflect.TypeOf(s.value)]; !ok {
		panic(fmt.Errorf("store for value of type %T is already destroyed", s.value))
	}
	delete(stores, reflect.TypeOf(s.value))
	s.value = nil
}

// Reducer is a mechanism through which a Store's value is changed. It
// is sent actions and is required to return the changed value according
// to the specified action. If the Reducer returns the old state of the
// Store, then it is considered that the action is not applicable for this
// Reducer.
type Reducer func(store *Store, action interface{}) interface{}

// InjectStore is a helper function that allows one to discover and access
// and arbitrary Store's value.
//
// The function uses the type of the referenced value by the target pointer
// to determine which Store should be used.
func InjectStore(target interface{}) {
	if target == nil {
		panic(fmt.Errorf("target cannot be nil"))
	}
	value := reflect.ValueOf(target)
	valueType := value.Type()
	if valueType.Kind() != reflect.Ptr {
		panic(fmt.Errorf("target %T must be a pointer", target))
	}
	if value.IsNil() {
		panic(fmt.Errorf("target pointer cannot be nil"))
	}
	targetRefType := valueType.Elem()

	store, ok := stores[targetRefType]
	if !ok {
		panic(fmt.Errorf("there is no store for values of type %s", targetRefType))
	}
	store.Inject(target)
}

// Dispatch is the mechanism through which Stores get modified. An action
// is provided and all Reducers attempt to process the action. Should a
// reducer change it's Store's value then all connected components are
// reconciled.
func Dispatch(action interface{}) {
	uiCtx.Schedule(func() {
		invalidateConnectedNodes := false
		for _, store := range stores {
			newValue := store.reducer(store, action)
			if newValue != store.value {
				store.value = newValue
				invalidateConnectedNodes = true
			}
		}
		if invalidateConnectedNodes {
			for _, node := range connectedNodes {
				if node.isValid() {
					node.reconcile(node.instance, node.scope)
				}
			}
		}
	})
}

// DataMapFunc controls how a connected component's data is calculated.
type DataMapFunc func(props Properties) interface{}

// CallbackMapFunc controls how a connected component's data is calculated.
type CallbackMapFunc func(props Properties) interface{}

// ConnectMapping is a configuration mechanism to wire a
// connected component to its delegate.
type ConnectMapping struct {

	// Data, if specified, controls the data to be passed to the
	// delegate component. If nil, then the original data is passed through.
	Data DataMapFunc

	// Callback, if specified, controls the callback data to be passed to the
	// delegate component. If nil, then the original callback data is passed
	// through.
	Callback CallbackMapFunc
}

// Connect is the mechanism through which a Component gets wired to
// the global Stores. A connected component will get invalidated when
// one of the Stores gets changed.
//
// The Connect method is used to wrap an existing Component and the
// mapping configuration can be used to adjust the delegate component's
// data and callback data based on the Store state.
func Connect(delegate Component, mapping ConnectMapping) Component {
	return Component{
		componentType: evaluateComponentType(),
		componentFunc: func(props Properties, scope Scope) Instance {
			Once(func() {
				addConnectedNode(renderCtx.node)
			})

			Defer(func() {
				removeConnectedNode(renderCtx.node)
			})

			data := props.data
			if mapping.Data != nil {
				data = mapping.Data(props)
			}
			callbackData := props.callbackData
			if mapping.Callback != nil {
				callbackData = mapping.Callback(props)
			}

			return delegate.componentFunc(Properties{
				data:         data,
				layoutData:   props.layoutData,
				callbackData: callbackData,
				children:     props.children,
			}, scope)
		},
	}
}

// StoreProviderData is the data necessary to instantiate a StoreProvider.
//
// Note that the StoreProvider caches the Stores it creates and a
// reconciliation with a different data will be ignored.
type StoreProviderData struct {

	// Entries contains the definition of all Stores that should be managed
	// by the given StoreProvider.
	Entries []StoreProviderEntry
}

// StoreProviderEntry represents a single Store instance.
type StoreProviderEntry struct {

	// Reducer specifies the Reducer that will be passed to CreateStore.
	Reducer Reducer

	// InitialValue specifies the initialValue that will be passed to CreateStore.
	InitialValue interface{}
}

// NewStoreProviderEntry is a helper function to quickly create a
// StoreProviderEntry value.
func NewStoreProviderEntry(reducer Reducer, initialValue interface{}) StoreProviderEntry {
	return StoreProviderEntry{
		Reducer:      reducer,
		InitialValue: initialValue,
	}
}

// StoreProvider is a convenience component that manages the lifecycle of
// a set of Stores.
//
// Using StoreProvider can be used as an optimization, since the lifecycle
// of the Stores is determined by the lifecycle of the given StoreProvider.
// If certain Store should be available only in the context of a given
// component hierarchy (e.g. a wizard dialog) it might make sense to
// wrap that hierarchy with such a StoreProvider.
//
// Should you go down this route, however, make sure that only nested
// components try to access Stores that are managed by this StoreProvider.
var StoreProvider = Define(func(props Properties, scope Scope) Instance {
	data := GetData[StoreProviderData](props)

	stores := UseState(func() []*Store {
		result := make([]*Store, len(data.Entries))
		for i, entry := range data.Entries {
			result[i] = CreateStore(entry.Reducer, entry.InitialValue)
		}
		return result
	}).Get()

	Defer(func() {
		for _, store := range stores {
			store.Destroy()
		}
	})

	return New(Element, func() {
		WithData(ElementData{
			Layout: ui.NewFillLayout(),
		})
		WithChildren(props.children)
	})
})

func addConnectedNode(node *componentNode) {
	connectedNodes = append(connectedNodes, node)
}

func removeConnectedNode(node *componentNode) {
	for i, connectedNode := range connectedNodes {
		if connectedNode == node {
			connectedNodes[i] = connectedNodes[len(connectedNodes)-1]
			connectedNodes = connectedNodes[:len(connectedNodes)-1]
		}
	}
}
