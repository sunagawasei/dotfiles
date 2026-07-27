package sourceinventory

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/sunagawasei/dotfiles/scripts/internal/verifycolors"
)

type ezaStyleExpectation struct {
	token    verifycolors.TokenRef
	modifier string
}

type ezaSourceSpec struct {
	path           string
	token          verifycolors.TokenRef
	modifier       string
	completionCode string
	extensions     []string
	source         string
}

var ezaThemeSchema = map[string]ezaStyleExpectation{
	"filekinds.normal":               {token: "foregrounds.main"},
	"filekinds.directory":            {token: "foregrounds.heading"},
	"filekinds.symlink":              {token: "semantic.keyword"},
	"filekinds.pipe":                 {token: "semantic.operator"},
	"filekinds.block_device":         {token: "foregrounds.dim"},
	"filekinds.char_device":          {token: "foregrounds.dim"},
	"filekinds.socket":               {token: "semantic.operator"},
	"filekinds.special":              {token: "teals.mid_bright"},
	"filekinds.executable":           {token: "semantic.success"},
	"filekinds.mount_point":          {token: "teals.mid_bright"},
	"perms.user_read":                {token: "foregrounds.dim"},
	"perms.user_write":               {token: "semantic.warning"},
	"perms.user_execute_file":        {token: "semantic.success"},
	"perms.user_execute_other":       {token: "semantic.success"},
	"perms.group_read":               {token: "foregrounds.dim"},
	"perms.group_write":              {token: "semantic.warning"},
	"perms.group_execute":            {token: "semantic.success"},
	"perms.other_read":               {token: "foregrounds.dim"},
	"perms.other_write":              {token: "semantic.warning"},
	"perms.other_execute":            {token: "semantic.success"},
	"perms.special_user_file":        {token: "purples.bright_purple"},
	"perms.special_other":            {token: "purples.bright_purple"},
	"perms.attribute":                {token: "foregrounds.subdued"},
	"size.major":                     {token: "foregrounds.dim"},
	"size.minor":                     {token: "foregrounds.dim"},
	"size.number_byte":               {token: "foregrounds.main"},
	"size.number_kilo":               {token: "foregrounds.main"},
	"size.number_mega":               {token: "foregrounds.main"},
	"size.number_giga":               {token: "foregrounds.main"},
	"size.number_huge":               {token: "foregrounds.main"},
	"size.unit_byte":                 {token: "foregrounds.subdued"},
	"size.unit_kilo":                 {token: "foregrounds.subdued"},
	"size.unit_mega":                 {token: "foregrounds.subdued"},
	"size.unit_giga":                 {token: "foregrounds.subdued"},
	"size.unit_huge":                 {token: "foregrounds.subdued"},
	"users.user_you":                 {token: "semantic.success"},
	"users.user_root":                {token: "semantic.error"},
	"users.user_other":               {token: "foregrounds.dim"},
	"users.group_yours":              {token: "semantic.success"},
	"users.group_other":              {token: "foregrounds.dim"},
	"users.group_root":               {token: "semantic.error"},
	"links.normal":                   {token: "foregrounds.dim"},
	"links.multi_link_file":          {token: "semantic.warning"},
	"git.new":                        {token: "semantic.success"},
	"git.modified":                   {token: "semantic.warning"},
	"git.deleted":                    {token: "semantic.error"},
	"git.renamed":                    {token: "semantic.keyword"},
	"git.typechange":                 {token: "semantic.keyword"},
	"git.ignored":                    {token: "foregrounds.subdued"},
	"git.conflicted":                 {token: "semantic.error"},
	"git_repo.branch_main":           {token: "foregrounds.heading"},
	"git_repo.branch_other":          {token: "semantic.keyword"},
	"git_repo.git_clean":             {token: "semantic.success"},
	"git_repo.git_dirty":             {token: "semantic.warning"},
	"security_context.none":          {token: "foregrounds.subdued"},
	"security_context.selinux.colon": {token: "foregrounds.subdued"},
	"security_context.selinux.user":  {token: "foregrounds.subdued"},
	"security_context.selinux.role":  {token: "foregrounds.subdued"},
	"security_context.selinux.typ":   {token: "foregrounds.subdued"},
	"security_context.selinux.range": {token: "foregrounds.subdued"},
	"file_type.image":                {token: "purples.lavender"},
	"file_type.video":                {token: "purples.muted_purple"},
	"file_type.music":                {token: "purples.bright_purple"},
	"file_type.lossless":             {token: "purples.bright_purple"},
	"file_type.crypto":               {token: "ansi.bright_yellow"},
	"file_type.document":             {token: "blues_slates.cloud_slate"},
	"file_type.compressed":           {token: "ansi.bright_red"},
	"file_type.temp":                 {token: "foregrounds.subdued"},
	"file_type.compiled":             {token: "foregrounds.dim"},
	"file_type.build":                {token: "teals.mid_bright"},
	"file_type.source":               {token: "semantic.type"},
	"punctuation":                    {token: "semantic.punctuation"},
	"date":                           {token: "foregrounds.dim"},
	"inode":                          {token: "foregrounds.subdued"},
	"blocks":                         {token: "foregrounds.subdued"},
	"header":                         {token: "foregrounds.heading", modifier: "bold"},
	"octal":                          {token: "foregrounds.subdued"},
	"flags":                          {token: "foregrounds.subdued"},
	"symlink_path":                   {token: "foregrounds.dim"},
	"control_char":                   {token: "semantic.error"},
	"broken_symlink":                 {token: "semantic.error"},
	"broken_path_overlay":            {token: "semantic.error"},
}

var (
	ezaSourcePathPattern     = regexp.MustCompile(`^[a-z_]+(?:\.[a-z_]+){0,2}$`)
	ezaSourceTokenPattern    = regexp.MustCompile(`^[a-z_]+\.[a-z_]+$`)
	ezaCompletionCodePattern = regexp.MustCompile(`^[A-Za-z]{2}$`)
	ezaExtensionPattern      = regexp.MustCompile(`^[a-z0-9+_-]+(?:\.[a-z0-9+_-]+)*$`)
)

func extractEzaTheme(root string, result *Result) error {
	const relative = "scripts/cmd/generate-colors/main.go"
	template, startLine, err := readRawStringConstant(sourcePath(root, relative), "ezaStyleSpecTemplate")
	if err != nil {
		return err
	}
	specs, err := parseEzaSourceSpecs(template, relative, startLine)
	if err != nil {
		return err
	}
	if err := validateGeneratedEzaTheme(sourcePath(root, "eza/theme.yml")); err != nil {
		return err
	}
	if err := validateZshCompletionConfig(sourcePath(root, "home-manager/zsh.nix"), specs); err != nil {
		return err
	}

	for _, spec := range specs {
		result.addPair(defaultTextPair(
			"eza.theme."+spec.path,
			spec.token,
			verifycolors.AmbientBackground(),
			[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
			spec.source,
		))
		if spec.completionCode != "" {
			result.addPair(defaultTextPair(
				"zsh.completion.class."+spec.completionCode,
				spec.token,
				verifycolors.AmbientBackground(),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				spec.source,
			))
		}
		for _, extension := range spec.extensions {
			result.addPair(defaultTextPair(
				"zsh.completion.extension."+strings.ReplaceAll(extension, ".", "_"),
				spec.token,
				verifycolors.AmbientBackground(),
				[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
				spec.source,
			))
		}
	}

	source := fmt.Sprintf("%s:%d", relative, startLine)
	result.addPair(defaultTextPair(
		"zsh.completion.menu-select",
		"core.selection_fg",
		verifycolors.TokenBackground("core.selection_bg"),
		[]verifycolors.RenderProfile{verifycolors.ProfileTruecolor},
		source,
	))
	result.addCoverageNote(verifycolors.CoverageNote{
		ID:     "zsh.completion.file-type-subset",
		Reason: "zsh list-colors statically covers only the generated extension set; other extensions, exact filenames such as README and Makefile, temporary-name suffixes, and source-adjacent compiled inference fall back to fi while eza keeps its internal file_type classification. LS_COLORS is unset by interactive zsh only, so future non-interactive scripts invoking eza must unset it themselves; no repository scripts currently invoke eza",
		Source: source,
	})
	return nil
}

func parseEzaSourceSpecs(template, relative string, startLine int) ([]ezaSourceSpec, error) {
	var specs []ezaSourceSpec
	seenPaths := make(map[string]bool)
	seenCodes := make(map[string]bool)
	seenExtensions := make(map[string]bool)
	for index, rawLine := range strings.Split(template, "\n") {
		lineNumber := startLine + index
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "|")
		if len(fields) != 5 {
			return nil, fmt.Errorf("%s:%d: eza style spec has %d fields, want 5", relative, lineNumber, len(fields))
		}
		for fieldIndex := range fields {
			fields[fieldIndex] = strings.TrimSpace(fields[fieldIndex])
		}
		if !ezaSourcePathPattern.MatchString(fields[0]) {
			return nil, fmt.Errorf("%s:%d: invalid eza theme path %q", relative, lineNumber, fields[0])
		}
		expectation, ok := ezaThemeSchema[fields[0]]
		if !ok {
			return nil, fmt.Errorf("%s:%d: unknown eza 0.23.4 theme path %q", relative, lineNumber, fields[0])
		}
		if seenPaths[fields[0]] {
			return nil, fmt.Errorf("%s:%d: duplicate eza theme path %q", relative, lineNumber, fields[0])
		}
		seenPaths[fields[0]] = true
		if !ezaSourceTokenPattern.MatchString(fields[1]) {
			return nil, fmt.Errorf("%s:%d: invalid eza theme token %q", relative, lineNumber, fields[1])
		}
		token := verifycolors.TokenRef(fields[1])
		if token != expectation.token {
			return nil, fmt.Errorf("%s:%d: eza theme path %q resolves to %q, want %q", relative, lineNumber, fields[0], token, expectation.token)
		}
		if fields[2] != expectation.modifier {
			return nil, fmt.Errorf("%s:%d: eza theme path %q modifier = %q, want %q", relative, lineNumber, fields[0], fields[2], expectation.modifier)
		}

		spec := ezaSourceSpec{
			path:           fields[0],
			token:          token,
			modifier:       fields[2],
			completionCode: fields[3],
			source:         fmt.Sprintf("%s:%d", relative, lineNumber),
		}
		if spec.completionCode != "" {
			if !ezaCompletionCodePattern.MatchString(spec.completionCode) {
				return nil, fmt.Errorf("%s:%d: invalid zsh completion code %q", relative, lineNumber, spec.completionCode)
			}
			if seenCodes[spec.completionCode] {
				return nil, fmt.Errorf("%s:%d: duplicate zsh completion code %q", relative, lineNumber, spec.completionCode)
			}
			seenCodes[spec.completionCode] = true
		}
		if fields[4] != "" {
			if !strings.HasPrefix(spec.path, "file_type.") {
				return nil, fmt.Errorf("%s:%d: extensions assigned to non-file_type path %q", relative, lineNumber, spec.path)
			}
			for _, extension := range strings.Split(fields[4], ",") {
				extension = strings.TrimSpace(extension)
				if !ezaExtensionPattern.MatchString(extension) {
					return nil, fmt.Errorf("%s:%d: invalid zsh completion extension %q", relative, lineNumber, extension)
				}
				if seenExtensions[extension] {
					return nil, fmt.Errorf("%s:%d: duplicate zsh completion extension %q", relative, lineNumber, extension)
				}
				seenExtensions[extension] = true
				spec.extensions = append(spec.extensions, extension)
			}
		}
		specs = append(specs, spec)
	}
	for path := range ezaThemeSchema {
		if !seenPaths[path] {
			return nil, fmt.Errorf("%s: eza 0.23.4 theme path %q was not found", relative, path)
		}
	}
	return specs, nil
}

func validateGeneratedEzaTheme(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	containers := make(map[string]bool)
	for stylePath := range ezaThemeSchema {
		parts := strings.Split(stylePath, ".")
		for index := 1; index < len(parts); index++ {
			containers[strings.Join(parts[:index], ".")] = true
		}
	}
	seen := make(map[string]bool)
	var parents []string
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		rawLine := scanner.Text()
		line := strings.TrimSpace(rawLine)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		leading := len(rawLine) - len(strings.TrimLeft(rawLine, " "))
		if leading%2 != 0 {
			return fmt.Errorf("%s:%d: eza theme indentation must use two spaces", path, lineNumber)
		}
		level := leading / 2
		if level > len(parents) {
			return fmt.Errorf("%s:%d: eza theme indentation skips a level", path, lineNumber)
		}
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 || parts[0] == "" {
			return fmt.Errorf("%s:%d: invalid eza theme entry %q", path, lineNumber, line)
		}
		key := parts[0]
		value := strings.TrimSpace(parts[1])
		parents = parents[:level]
		fullParts := append(append([]string{}, parents...), key)
		fullPath := strings.Join(fullParts, ".")
		if seen[fullPath] {
			return fmt.Errorf("%s:%d: duplicate eza theme key %q", path, lineNumber, fullPath)
		}
		seen[fullPath] = true

		if value == "" {
			if !containers[fullPath] {
				if _, ok := ezaThemeSchema[fullPath]; !ok {
					return fmt.Errorf("%s:%d: unknown eza 0.23.4 theme key or hierarchy %q", path, lineNumber, fullPath)
				}
			}
			parents = fullParts
			continue
		}

		switch {
		case fullPath == "colourful" && value == "true":
		case fullPath == "filenames" && value == "{}":
		case fullPath == "extensions" && value == "{}":
		case key == "foreground":
			stylePath := strings.Join(parents, ".")
			if _, ok := ezaThemeSchema[stylePath]; !ok {
				return fmt.Errorf("%s:%d: foreground belongs to unknown eza theme path %q", path, lineNumber, stylePath)
			}
			if !regexp.MustCompile(`^"#[0-9A-F]{6}"$`).MatchString(value) {
				return fmt.Errorf("%s:%d: invalid generated eza foreground %q", path, lineNumber, value)
			}
		case key == "is_bold":
			stylePath := strings.Join(parents, ".")
			if ezaThemeSchema[stylePath].modifier != "bold" || value != "true" {
				return fmt.Errorf("%s:%d: unexpected eza style modifier %q at %q", path, lineNumber, value, stylePath)
			}
		default:
			return fmt.Errorf("%s:%d: unknown or invalid eza 0.23.4 theme key %q", path, lineNumber, fullPath)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	for _, required := range []string{"colourful", "filenames", "extensions"} {
		if !seen[required] {
			return fmt.Errorf("%s: required eza theme key %q was not found", path, required)
		}
	}
	for stylePath, expectation := range ezaThemeSchema {
		if !seen[stylePath+".foreground"] {
			return fmt.Errorf("%s: eza theme foreground %q was not found", path, stylePath)
		}
		modifierPath := stylePath + ".is_bold"
		if expectation.modifier == "bold" && !seen[modifierPath] {
			return fmt.Errorf("%s: eza theme modifier %q was not found", path, modifierPath)
		}
		if expectation.modifier == "" && seen[modifierPath] {
			return fmt.Errorf("%s: unexpected eza theme modifier %q", path, modifierPath)
		}
	}
	return nil
}

func validateZshCompletionConfig(path string, specs []ezaSourceSpec) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(data)
	if strings.Contains(content, "export LS_COLORS") {
		return fmt.Errorf("%s: LS_COLORS must not be exported", path)
	}
	for _, required := range []string{
		"        unset LS_COLORS",
		"        typeset -ga ZSH_COMPLETION_COLORS=(",
		`zstyle ':completion:*' list-colors "''${ZSH_COMPLETION_COLORS[@]}"`,
	} {
		if strings.Count(content, required) != 1 {
			return fmt.Errorf("%s: expected exactly one %q", path, required)
		}
	}

	expectedEntries := map[string]bool{"ma": true}
	for _, spec := range specs {
		if spec.completionCode != "" {
			expectedEntries[spec.completionCode] = true
		}
		for _, extension := range spec.extensions {
			expectedEntries["*."+extension] = true
		}
	}
	entryPattern := regexp.MustCompile(`(?m)^[ \t]+'([^=\r\n]+)=([^'\r\n]+)'$`)
	seenEntries := make(map[string]bool)
	for _, match := range entryPattern.FindAllStringSubmatch(content, -1) {
		key, style := match[1], match[2]
		if !expectedEntries[key] {
			return fmt.Errorf("%s: unknown generated zsh completion entry %q", path, key)
		}
		if seenEntries[key] {
			return fmt.Errorf("%s: duplicate generated zsh completion entry %q", path, key)
		}
		seenEntries[key] = true
		pattern := regexp.MustCompile(`^38;2;(?:\d+;){2}\d+$`)
		if key == "ma" {
			pattern = regexp.MustCompile(`^48;2;(?:\d+;){2}\d+;38;2;(?:\d+;){2}\d+$`)
		}
		if !pattern.MatchString(style) {
			return fmt.Errorf("%s: invalid truecolor SGR for zsh completion entry %q", path, key)
		}
	}
	for key := range expectedEntries {
		if !seenEntries[key] {
			return fmt.Errorf("%s: generated zsh completion entry %q was not found", path, key)
		}
	}
	return nil
}

func readRawStringConstant(path, name string) (string, int, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer file.Close()

	prefix := "const " + name + " = `"
	var builder strings.Builder
	found := false
	startLine := 0
	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		if !found {
			if strings.HasPrefix(line, prefix) {
				found = true
				startLine = lineNumber
				builder.WriteString(strings.TrimPrefix(line, prefix))
				builder.WriteByte('\n')
			}
			continue
		}
		if line == "`" {
			return builder.String(), startLine, nil
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return "", 0, err
	}
	if !found {
		return "", 0, fmt.Errorf("%s was not found", name)
	}
	return "", 0, fmt.Errorf("%s is not terminated", name)
}
