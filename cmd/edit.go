package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const bodyStartLine uint8 = 7

var (
	bodyFlag string
)

var editCmd = &cobra.Command{
	Use:   "edit <id|slug>",
	Short: "Open a task in $EDITOR",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		path, err := FindByRef(args[0])
		if err != nil {
			return err
		}

		if bodyFlag != "" {
			tmp := path + ".tmp"

			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()

			tmpFile, err := os.Create(tmp)
			if err != nil {
				return err
			}
			defer tmpFile.Close()

			scanner := bufio.NewScanner(file)
			writer := bufio.NewWriter(tmpFile)
			lineCount := uint8(0)

			for scanner.Scan() {
				lineCount++

				fmt.Fprintln(writer, scanner.Text())

				if lineCount == bodyStartLine {
					fmt.Fprint(writer, "\n")

					for newLine := range strings.Lines(bodyFlag) {
						fmt.Fprintln(writer, newLine)
					}
					break
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}

			writer.Flush()

			return os.Rename(tmp, path)
		}

		editor := viper.GetString("EDITOR")
		if editor == "" {
			editor = "vi"
		}
		c := exec.Command(editor, path)
		c.Stdin = os.Stdin
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr

		if err := c.Run(); err != nil {
			return fmt.Errorf("editor: %w", err)
		}
		return nil
	},
}

func init() {
	editCmd.Flags().StringVar(&bodyFlag, "body", "", "The body of the task after the header")

	rootCmd.AddCommand(editCmd)
}
