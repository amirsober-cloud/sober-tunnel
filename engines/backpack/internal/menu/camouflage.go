package menu

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/backpack/backpack/internal/mimicmgr"
	"github.com/backpack/backpack/internal/tui"
)

// camouflageMenu shares the panel's instance manager and native setup tools.
func camouflageMenu() {
	for {
		tui.Clear()
		choice := tui.ChooseOpt("Camouflage / Hedioum", []tui.Option{
			{Title: "Setup Foreign", Desc: "exit server, native mimic and persona settings"},
			{Title: "Setup Iran", Desc: "hub, foreign endpoints, SOCKS and TUN"},
			{Title: "Manage Instances", Desc: "edit, start, stop, restart, logs and pairing"},
			{Title: "Import Native JSON", Desc: "validate and register a Hedioum configuration"},
		})
		switch choice {
		case 0, 1:
			name := tui.Prompt("Instance name: ")
			if err := mimicmgr.CheckName(name); err != nil {
				camouflageResult(err)
				continue
			}
			if _, err := mimicmgr.Default.Read(name); err == nil {
				camouflageResult(fmt.Errorf("instance already exists; use Manage Instances"))
				continue
			} else if !os.IsNotExist(err) {
				camouflageResult(err)
				continue
			}
			command := "setup-foreign"
			if choice == 1 {
				command = "setup-iran"
			}
			camouflagePrepare(name, command)
		case 2:
			camouflageInstances()
		case 3:
			name := tui.Prompt("Instance name: ")
			if err := mimicmgr.CheckName(name); err != nil {
				camouflageResult(err)
				continue
			}
			path := tui.Prompt("Native JSON file path: ")
			data, err := os.ReadFile(path)
			if err == nil && tui.Confirm("Save this configuration (existing instance must be stopped)", false) {
				err = mimicmgr.Default.Save(context.Background(), name, data)
			}
			camouflageResult(err)
		default:
			return
		}
	}
}

func camouflageInstances() {
	for {
		items, err := mimicmgr.Default.List(context.Background())
		if err != nil {
			camouflageResult(err)
			return
		}
		if len(items) == 0 {
			tui.Info("No camouflage instances. Choose Setup Foreign or Setup Iran first.")
			tui.PressEnter()
			return
		}
		opts := make([]tui.Option, len(items))
		for i, item := range items {
			state := "stopped"
			if item.Active {
				state = "running"
			}
			opts[i] = tui.Option{Title: item.Name, Desc: item.Role + " - " + state}
		}
		idx := tui.ChooseOpt("Camouflage Instances", opts)
		if idx < 0 {
			return
		}
		camouflageInstance(items[idx])
	}
}

func camouflageInstance(item mimicmgr.Instance) {
	for {
		choice := tui.ChooseOpt("Camouflage: "+item.Name, []tui.Option{
			{Title: "Edit Native Settings", Desc: "same setup/edit flags as Hedioum; stop before saving"},
			{Title: "Add Node", Desc: "Iran hub only"},
			{Title: "Edit Node", Desc: "Iran hub only"},
			{Title: "Remove Node", Desc: "Iran hub only"},
			{Title: "Start"}, {Title: "Stop"}, {Title: "Restart"},
			{Title: "Logs", Desc: "last 150 journal lines"},
			{Title: "Show Native JSON", Desc: "contains credentials"},
			{Title: "Pairing Token", Desc: "foreign exit only; contains credentials"},
			{Title: "Delete", Desc: "stop and remove this instance"},
			{Title: "Diagnostics", Desc: "probe, connection test, speedtest and exit IP"},
		})
		switch choice {
		case 0:
			command := "edit-foreign"
			if item.Role == "iran" {
				command = "edit-node"
			}
			camouflagePrepare(item.Name, command)
		case 1, 2, 3:
			if item.Role != "iran" {
				camouflageResult(fmt.Errorf("node commands require an Iran hub"))
				continue
			}
			camouflagePrepare(item.Name, []string{"add-node", "edit-node", "remove-node"}[choice-1])
		case 4, 5, 6:
			camouflageResult(mimicmgr.Default.Action(context.Background(), item.Name, []string{"start", "stop", "restart"}[choice-4]))
		case 7:
			output, err := mimicmgr.Default.Logs(context.Background(), item.Name)
			fmt.Println(output)
			camouflageResult(err)
		case 8:
			data, err := mimicmgr.Default.Read(item.Name)
			if err == nil {
				fmt.Println(string(data))
			}
			camouflageResult(err)
		case 9:
			if item.Role != "foreign" {
				camouflageResult(fmt.Errorf("pairing requires a foreign exit"))
				continue
			}
			camouflagePrepare(item.Name, "pairing")
		case 11:
			camouflageDiagnostics(item.Name)
		case 10:
			if tui.Confirm("Delete "+item.Name, false) {
				err := mimicmgr.Default.Action(context.Background(), item.Name, "delete")
				camouflageResult(err)
				if err == nil {
					return
				}
			}
		default:
			return
		}
	}
}

// Native flag metadata keeps new camouflage options available without duplicating
// the upstream configuration logic. Only explicitly chosen flags are sent.
func camouflagePrepare(name, command string) {
	flags := mimicmgr.ToolOptions()[command]
	values := map[string]string{}
	for {
		opts := []tui.Option{{Title: "Preview", Desc: "run native validation and review the result"}}
		for _, flag := range flags {
			value := flag.Default
			if chosen, ok := values[flag.Name]; ok {
				value = chosen
			}
			if strings.Contains(flag.Name, "token") && value != "" {
				value = "[hidden]"
			}
			opts = append(opts, tui.Option{Title: "--" + flag.Name + " = " + value, Desc: flag.Help})
		}
		idx := tui.ChooseOpt(command+": "+name, opts)
		if idx < 0 {
			return
		}
		if idx == 0 {
			break
		}
		flag := flags[idx-1]
		value, ok := tui.PromptOrEnd("Value (empty uses native default): ")
		if !ok {
			return
		}
		if value == "" {
			delete(values, flag.Name)
		} else {
			values[flag.Name] = value
		}
	}
	result, err := mimicmgr.Default.Prepare(context.Background(), mimicmgr.ToolRequest{Name: name, Command: command, Flags: values})
	if err != nil {
		camouflageResult(err)
		return
	}
	preview, ok := result.(map[string]any)
	if !ok {
		camouflageResult(fmt.Errorf("unexpected native preview result"))
		return
	}
	fmt.Println(preview["output"])
	if command == "pairing" {
		tui.PressEnter()
		return
	}
	data, ok := preview["config"].(json.RawMessage)
	if !ok {
		camouflageResult(fmt.Errorf("native preview has no configuration"))
		return
	}
	fmt.Println(string(data))
	if tui.Confirm("Save configuration (does not start the tunnel)", false) {
		camouflageResult(mimicmgr.Default.Save(context.Background(), name, data))
	}
}

func camouflageResult(err error) {
	if err != nil {
		tui.Error(err.Error())
	}
	tui.PressEnter()
}

func camouflageDiagnostics(name string) {
	choice := tui.ChooseOpt("Native Diagnostics", []tui.Option{
		{Title: "Probe", Desc: "authenticate each endpoint"},
		{Title: "Connection Test", Desc: "exit IP, websites and DNS"},
		{Title: "Speedtest", Desc: "measure tunnel throughput"},
		{Title: "Check IP", Desc: "exit IP reputation"},
	})
	if choice < 0 {
		return
	}
	command := []string{"probe", "test", "speedtest", "check-ip"}[choice]
	flags := map[string]string{}
	if command != "check-ip" {
		node, ok := tui.PromptOrEnd("Node alias (empty uses native default): ")
		if !ok {
			return
		}
		if node != "" {
			flags["node"] = node
		}
	}
	job, err := mimicmgr.Default.StartDiagnostic(mimicmgr.ToolRequest{Name: name, Command: command, Flags: flags})
	if err != nil {
		camouflageResult(err)
		return
	}
	for {
		job, err = mimicmgr.Default.Diagnostic(name, job.ID)
		if err != nil {
			camouflageResult(err)
			return
		}
		tui.Clear()
		fmt.Println(job.Output)
		if job.Done {
			if job.Error != "" {
				tui.Error(job.Error)
			}
			tui.PressEnter()
			return
		}
		if _, ok := tui.PromptOrEnd("Running: Enter to refresh, EOF to return: "); !ok {
			return
		}
	}
}
