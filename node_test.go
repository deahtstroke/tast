package tast

import (
	"reflect"
	"testing"
)

func Test_Table(t *testing.T) {
	tests := map[string]struct {
		doc       *Document
		searchKey string
		wantErr   bool
	}{
		"Single key segment should be found": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"key"},
						},
					},
				},
			},
			searchKey: "key",
		},
		"Multiple key segments should be found": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"foo", "bar", "bez"},
						},
					},
				},
			},
			searchKey: "foo.bar.bez",
		},
		"Key not found": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"foo", "bar", "bez"},
						},
					},
				},
			},
			searchKey: "foo.bar.be",
			wantErr:   true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			_, ok := tt.doc.Table(tt.searchKey)
			if tt.wantErr {
				if ok {
					t.Fatalf("Didn't wantt able, got one?")
				}
				return
			}

			if !ok {
				t.Fatalf("Want table, got none")
			}
		})
	}
}

func Test_Set(t *testing.T) {
	tests := map[string]struct {
		doc      *Document
		key      string
		table    string
		value    any
		wantType Node
		wantErr  bool
	}{
		"Simple key string": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo"},
								},
								value: &StringNode{
									value: "bar",
								},
							},
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"hello"},
								},
								value: &StringNode{
									value: "world",
								},
							},
						},
					},
				},
			},
			table:    "table",
			key:      "foo",
			value:    int64(1),
			wantType: &IntegerNode{}, // Just need the type, value can be ignored
			wantErr:  false,
		},
		"dotted key string": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo", "bar"},
								},
								value: &StringNode{
									value: "hello world!",
								},
							},
						},
					},
				},
			},
			table:    "table",
			key:      "foo.bar",
			value:    int64(1),
			wantType: &IntegerNode{}, // Just need the type, value can be ignored
			wantErr:  false,
		},
		"Simple Key replaced by string node": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo", "bar"},
								},
								value: &StringNode{
									value: "hello world!",
								},
							},
						},
					},
				},
			},
			table:    "table",
			key:      "foo.bar",
			value:    string("Hello!"),
			wantType: &StringNode{},
			wantErr:  false,
		},
		"Simple Key replaced by boolean node": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo", "bar"},
								},
								value: &StringNode{
									value: "hello world!",
								},
							},
						},
					},
				},
			},
			table:    "table",
			key:      "foo.bar",
			value:    bool(true),
			wantType: &BooleanNode{},
			wantErr:  false,
		},
		"Simple key not found should return false": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo"},
								},
								value: &StringNode{
									value: "hello world!",
								},
							},
						},
					},
				},
			},
			table:   "table",
			key:     "fo",
			value:   bool(true),
			wantErr: true,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table, ok := tt.doc.Table(tt.table)
			if !ok {
				t.Fatalf("Not expecting error, got one")
			}

			err := table.Set(tt.key, tt.value)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("Wanted error, got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("Not expecting error, got %v", err)
			}

			got, ok := table.FindKey(tt.key)
			if !ok {
				t.Fatalf("key %q not found after Set", tt.key)
			}

			gotType := reflect.TypeOf(got.value)
			wantType := reflect.TypeOf(tt.wantType)

			if gotType != wantType {
				t.Fatalf("expected node type %v, got %v", wantType, gotType)
			}
		})
	}
}

func Test_Delete(t *testing.T) {
	tests := map[string]struct {
		doc       *Document
		tableKey  string
		deleteKey string
		wantErr   bool
	}{
		"Delete should find simple keys": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"key"},
								},
								value: &StringNode{value: "hello world!"},
							},
						},
					},
				},
			},
			tableKey:  "table",
			deleteKey: "key",
			wantErr:   false,
		},
		"Delete should find dotted keys": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo", "bar"},
								},
								value: &StringNode{value: "hello world!"},
							},
						},
					},
				},
			},
			tableKey:  "table",
			deleteKey: "foo.bar",
			wantErr:   false,
		},
		"Delete should fail on key not-found": {
			doc: &Document{
				content: []Node{
					&TableNode{
						key: &KeyNode{
							segments: []string{"table"},
						},
						children: []Node{
							&KeyValueNode{
								key: &KeyNode{
									segments: []string{"foo"},
								},
								value: &StringNode{value: "hello world!"},
							},
						},
					},
				},
			},
			tableKey:  "table",
			deleteKey: "bar",
			wantErr:   true,
		},
	}

	for test, tt := range tests {
		t.Run(test, func(t *testing.T) {
			table, ok := tt.doc.Table(tt.tableKey)
			if !ok {
				t.Fatalf("Expecting table %s, got none", tt.deleteKey)
			}

			ok = table.Delete(tt.deleteKey)

			if tt.wantErr {
				if ok {
					t.Fatalf("Want error, got none")
				}
				return
			}

			if !ok {
				t.Fatalf("Expected successful deletion unsuccessful")
			}
		})
	}
}
