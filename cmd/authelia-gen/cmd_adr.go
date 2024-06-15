package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

func newADRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "adr",
		Short: "Generate an Architecture Decision Record",

		DisableAutoGenTag: true,
	}

	cmd.AddCommand(newADRAddCmd())

	return cmd
}

func newADRAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add an Architecture Decision Record",
		RunE:  adrAddRunE,

		DisableAutoGenTag: true,
	}

	cmd.Flags().String("title", "", "sets the title of the record")
	cmd.Flags().String("status", "", "sets the status of the record")
	cmd.Flags().String("context", "", "sets the context of the record")
	cmd.Flags().String("decision", "", "sets the decision of the record")
	cmd.Flags().String("consequences", "", "sets the consequences of the record")

	return cmd
}

func adrAddRunE(cmd *cobra.Command, args []string) (err error) {
	var adrs string

	if adrs, err = getPFlagPath(cmd.Flags(), cmdFlagRoot, cmdFlagDocs, cmdFlagDocsContent, cmdFlagDocsADR); err != nil {
		return err
	}

	c := filepath.Join(adrs, ".adr.config.json")

	var raw []byte

	if raw, err = os.ReadFile(c); err != nil {
		return fmt.Errorf("error opening adr config: %w", err)
	}

	var config ArchitectureDesignRecordConfig

	if err = json.Unmarshal(raw, &config); err != nil {
		return fmt.Errorf("error parsing adr config: %w", err)
	}

	data := &ArchitectureDesignRecordTmpl{
		ADR:       config.NextID,
		Weight:    1000 + config.NextID,
		Date:      time.Now().Format(dateFmtYAML),
		DateHuman: time.Now().Format("January 2, 2006"),
	}

	if data.Title, err = cmd.Flags().GetString("title"); err != nil {
		return err
	}

	if data.Status, err = cmd.Flags().GetString("status"); err != nil {
		return err
	}

	if data.Context, err = cmd.Flags().GetString("context"); err != nil {
		return err
	}

	if data.Decision, err = cmd.Flags().GetString("decision"); err != nil {
		return err
	}

	if data.Consequences, err = cmd.Flags().GetString("consequences"); err != nil {
		return err
	}

	var f *os.File

	if f, err = os.Create(filepath.Join(adrs, fmt.Sprintf("%d-%s.md", data.ADR, strings.ReplaceAll(strings.ToLower(data.Title), " ", "-")))); err != nil {
		return fmt.Errorf("error opening file for adr: %w", err)
	}

	defer f.Close()

	if err = tmplADR.Execute(f, data); err != nil {
		return fmt.Errorf("error executing template: %w", err)
	}

	config.NextID += 1

	if raw, err = json.Marshal(config); err != nil {
		return fmt.Errorf("error serializing config: %w", err)
	}

	if err = os.WriteFile(c, raw, 0600); err != nil {
		return fmt.Errorf("error writing config: %w", err)
	}

	return nil
}

type ArchitectureDesignRecordConfig struct {
	NextID int `json:"next_id"`
}

type ArchitectureDesignRecordTmpl struct {
	ADR          int
	Weight       int
	Date         string
	DateHuman    string
	Title        string
	Status       string
	Context      string
	Decision     string
	Consequences string
}
