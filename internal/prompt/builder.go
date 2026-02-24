// Package prompt builds system prompts for Flynn.
package prompt

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Mode string

const (
	ModeFull    Mode = "full"
	ModeMinimal Mode = "minimal"
)

type Builder struct {
	Mode          Mode
	MaxFileChars  int
	MaxTotalChars int
	Workspace     string
	Timezone      string
}

type SystemContext struct {
	Tooling   string
	Safety    string
	Memory    string
	Runtime   string
	Workspace string
	Bootstrap []BootstrapFile
}

type BootstrapFile struct {
	Name    string
	Content string
	Trunc   bool
	Missing bool
}

func NewBuilder(mode Mode) *Builder {
	return &Builder{
		Mode:          mode,
		MaxFileChars:  2000,
		MaxTotalChars: 8000,
	}
}

func (b *Builder) BuildSystemPrompt(ctx SystemContext) string {
	var sections []string
	sections = append(sections, "Identity:\nYou are Flynn, a local-first personal OS agent. Be concise and action-oriented.")
	sections = append(sections, "Tooling:\n"+nonEmpty(ctx.Tooling, "None."))
	sections = append(sections, "Safety:\n"+nonEmpty(ctx.Safety, "Confirm destructive or admin-level actions."))

	if b.Mode == ModeFull {
		sections = append(sections, "Memory Policy:\n"+nonEmpty(ctx.Memory, "Store only durable user facts and personal actions. Ignore transient chat."))
		sections = append(sections, "Workspace:\n"+nonEmpty(ctx.Workspace, b.workspaceLine()))
		sections = append(sections, "Runtime:\n"+nonEmpty(ctx.Runtime, b.runtimeLine()))
		sections = append(sections, "Current Date & Time:\n"+b.timeLine())
		bootstrap := b.bootstrapSection(ctx.Bootstrap)
		if bootstrap != "" {
			sections = append(sections, bootstrap)
		}
	} else {
		sections = append(sections, "Workspace:\n"+nonEmpty(ctx.Workspace, b.workspaceLine()))
		sections = append(sections, "Runtime:\n"+nonEmpty(ctx.Runtime, b.runtimeLine()))
	}

	return strings.Join(sections, "\n\n")
}

func (b *Builder) workspaceLine() string {
	if b.Workspace != "" {
		return b.Workspace
	}
	wd, _ := os.Getwd()
	if wd == "" {
		return "unknown"
	}
	return wd
}

func (b *Builder) runtimeLine() string {
	return fmt.Sprintf("%s/%s go=%s", runtime.GOOS, runtime.GOARCH, runtime.Version())
}

func (b *Builder) timeLine() string {
	if b.Timezone != "" {
		return fmt.Sprintf("Timezone: %s", b.Timezone)
	}
	return fmt.Sprintf("Timezone: %s", time.Now().Location())
}

func (b *Builder) bootstrapSection(files []BootstrapFile) string {
	if len(files) == 0 {
		return ""
	}
	var bld strings.Builder
	bld.WriteString("Project Context:\n")
	for _, f := range files {
		if f.Missing {
			bld.WriteString(fmt.Sprintf("- %s: (missing)\n", f.Name))
			continue
		}
		bld.WriteString(fmt.Sprintf("- %s:\n", f.Name))
		bld.WriteString(f.Content)
		if !strings.HasSuffix(f.Content, "\n") {
			bld.WriteString("\n")
		}
		if f.Trunc {
			bld.WriteString("[truncated]\n")
		}
	}
	return strings.TrimSpace(bld.String())
}

func (b *Builder) LoadBootstrapFiles(paths []string) []BootstrapFile {
	var out []BootstrapFile
	total := 0
	for _, p := range paths {
		name := filepath.Base(p)
		data, err := os.ReadFile(p)
		if err != nil {
			out = append(out, BootstrapFile{Name: name, Missing: true})
			continue
		}
		content := string(data)
		trunc := false
		if b.MaxFileChars > 0 && len(content) > b.MaxFileChars {
			content = content[:b.MaxFileChars]
			trunc = true
		}
		if b.MaxTotalChars > 0 && total+len(content) > b.MaxTotalChars {
			remain := b.MaxTotalChars - total
			if remain < 0 {
				remain = 0
			}
			content = content[:remain]
			trunc = true
		}
		total += len(content)
		out = append(out, BootstrapFile{Name: name, Content: content, Trunc: trunc})
		if b.MaxTotalChars > 0 && total >= b.MaxTotalChars {
			break
		}
	}
	return out
}

func nonEmpty(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}
