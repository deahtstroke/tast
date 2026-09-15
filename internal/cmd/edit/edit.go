package edit

import (
	"fmt"
	"os"
	"strings"

	"github.com/deahtstroke/tast"
	"github.com/spf13/cobra"
)

type editOpts struct {
	File string
}

func NewEditCommand() *cobra.Command {
	var opts editOpts
	cmd := &cobra.Command{
		Use:   "edit",
		Short: "Edit a TOML file via key-value arguments",
		Args:  cobra.MinimumNArgs(2),
		Long: `Edits a TOML file by specifying via key-value arguments which
values correspond to what in the document.

Example:
	tast edit --file config.toml datasource.host postgres datasource.port 5432
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			b, err := os.ReadFile(opts.File)
			if err != nil {
				return err
			}

			doc, err := tast.ParseBytes(b)
			if err != nil {
				return err
			}

			for i := 0; i < len(args); i += 2 {
				key := args[i]
				value := args[i+1]

				segs := strings.Split(key, ".")
				if len(segs) > 1 {
					key = strings.Join(segs[:len(segs)-1], ".")
					node, ok := doc.Table(key)
					if !ok {
						return fmt.Errorf("Unable to find key %s", key)
					}

					node.Set(segs[len(segs)-1], value)
				} else {
					node, ok := doc.FindKey(key)
					if !ok {
						return fmt.Errorf("Unable to find key %s", key)
					}

					node.Set(value)
				}
			}

			return doc.Save(opts.File)
		},
	}

	flags := cmd.Flags()
	flags.StringVarP(&opts.File, "file", "f", "", "Source file to edit")

	return cmd
}
