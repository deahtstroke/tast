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
				Content: []Node{
					&TableNode{
						Key: &KeyNode{
							Segments: []string{"key"},
						},
					},
				},
			},
			searchKey: "key",
		},
		"Multiple key segments should be found": {
			doc: &Document{
				Content: []Node{
					&TableNode{
						Key: &KeyNode{
							Segments: []string{"foo", "bar", "bez"},
						},
					},
				},
			},
			searchKey: "foo.bar.bez",
		},
		"Key not found": {
			doc: &Document{
				Content: []Node{
					&TableNode{
						Key: &KeyNode{
							Segments: []string{"foo", "bar", "bez"},
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
				Content: []Node{
					&TableNode{
						Key: &KeyNode{
							Segments: []string{"table"},
						},
						Children: []Node{
							&KeyValueNode{
								Key: &KeyNode{
									Segments: []string{"foo"},
								},
								Value: &StringNode{
									Value: "bar",
								},
							},
							&KeyValueNode{
								Key: &KeyNode{
									Segments: []string{"hello"},
								},
								Value: &StringNode{
									Value: "world",
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
				Content: []Node{
					&TableNode{
						Key: &KeyNode{
							Segments: []string{"table"},
						},
						Children: []Node{
							&KeyValueNode{
								Key: &KeyNode{
									Segments: []string{"foo.bar"},
								},
								Value: &StringNode{
									Value: "hello world!",
								},
							},
							&KeyValueNode{
								Key: &KeyNode{
									Segments: []string{"hello"},
								},
								Value: &StringNode{
									Value: "world",
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
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			table, ok := tt.doc.Table(tt.table)
			if tt.wantErr {
				if ok {
					t.Fatalf("Wanted error, got none")
				}
				return
			}

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

			gotType := reflect.TypeOf(got.Value)
			wantType := reflect.TypeOf(tt.wantType)

			if gotType != wantType {
				t.Fatalf("expected node type %v, got %v", wantType, gotType)
			}
		})
	}
}
