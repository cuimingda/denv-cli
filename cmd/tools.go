package cmd

var supportedTools = []string{
    "php",
    "python",
    "node",
    "go",
}

func IsSupportedTool(name string) bool {
    for _, item := range supportedTools {
        if item == name {
            return true
        }
    }
    return false
}

func SupportedTools() []string {
    out := make([]string, len(supportedTools))
    copy(out, supportedTools)
    return out
}
