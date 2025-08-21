# Go Best Practices - Quick Reference

## 🚀 **Essential Go Rules (Keep It Simple!)**

### 1. **Naming Conventions**
```go
// ✅ Good - Clear, descriptive
func GetPipelines() []Pipeline
var userToken string
type GitLabClient struct{}

// ❌ Bad - Unclear, abbreviated  
func GetPipes() []Pipe
var ut string
type GLClient struct{}
```

### 2. **Error Handling (The Go Way)**
```go
// ✅ Always check errors immediately
result, err := doSomething()
if err != nil {
    return fmt.Errorf("failed to do something: %w", err)
}

// ❌ Never ignore errors
result, _ := doSomething() // DON'T DO THIS!
```

### 3. **Package Structure**
```
project/
├── cmd/           # Main applications
├── internal/      # Private code (can't be imported)
├── pkg/          # Public libraries (if any)
└── docs/         # Documentation
```

### 4. **Function Design**
```go
// ✅ Good - Single responsibility, clear return
func ValidateToken(token string) error {
    if token == "" {
        return errors.New("token cannot be empty")
    }
    return nil
}

// ❌ Bad - Does too many things
func ValidateAndProcessTokenAndSaveToDatabase(token string) (bool, string, error)
```

### 5. **Struct Design**
```go
// ✅ Good - Exported fields when needed
type Pipeline struct {
    ID     int    `json:"id"`
    Status string `json:"status"`
    Branch string `json:"ref"`
}

// ✅ Good - Unexported fields for internal use
type client struct {
    token string
    url   string
}
```

## 🛠️ **Code Organization**

### **File Naming**
- `snake_case.go` - Use underscores
- `client.go` - Simple, descriptive names
- `pipeline_test.go` - Tests end with `_test.go`

### **Import Groups**
```go
import (
    // 1. Standard library
    "fmt"
    "os"
    
    // 2. Third-party packages
    "github.com/charmbracelet/bubbletea"
    
    // 3. Your project packages
    "github.com/rkristelijn/glab-tui/internal/core"
)
```

### **Interface Design**
```go
// ✅ Small, focused interfaces
type PipelineGetter interface {
    GetPipelines() ([]Pipeline, error)
}

// ✅ Name interfaces with -er suffix
type Runner interface {
    Run() error
}
```

## 🎯 **Common Patterns**

### **Constructor Pattern**
```go
// ✅ Use New* functions for constructors
func NewGitLabClient(token string) *GitLabClient {
    return &GitLabClient{
        token: token,
        client: &http.Client{},
    }
}
```

### **Context Usage**
```go
// ✅ Pass context as first parameter
func GetPipelines(ctx context.Context, projectID int) ([]Pipeline, error) {
    // Use ctx for cancellation, timeouts
}
```

### **JSON Handling**
```go
// ✅ Use struct tags for JSON
type Pipeline struct {
    ID     int    `json:"id"`
    Status string `json:"status"`
    WebURL string `json:"web_url"`
}
```

## ⚡ **Performance Tips**

### **String Building**
```go
// ✅ Use strings.Builder for multiple concatenations
var builder strings.Builder
builder.WriteString("Hello")
builder.WriteString(" World")
result := builder.String()

// ❌ Avoid repeated string concatenation
result := "Hello" + " " + "World" // OK for small cases
```

### **Slice Operations**
```go
// ✅ Pre-allocate slices when size is known
pipelines := make([]Pipeline, 0, expectedSize)

// ✅ Use append correctly
pipelines = append(pipelines, newPipeline)
```

## 🧪 **Testing Basics**

### **Test Function Names**
```go
// ✅ Clear test names
func TestGetPipelines_Success(t *testing.T) {}
func TestGetPipelines_InvalidToken(t *testing.T) {}
```

### **Table-Driven Tests**
```go
func TestValidateToken(t *testing.T) {
    tests := []struct {
        name    string
        token   string
        wantErr bool
    }{
        {"valid token", "glpat-123", false},
        {"empty token", "", true},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidateToken(tt.token)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidateToken() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }
}
```

## 🔧 **Tools & Commands**

### **Essential Commands**
```bash
go fmt ./...        # Format all code
go vet ./...        # Static analysis
go mod tidy         # Clean up dependencies
go build .          # Build the project
go test ./...       # Run all tests
go run main.go      # Run directly
```

### **Useful Tools**
```bash
# Install these for better development
go install golang.org/x/tools/cmd/goimports@latest  # Better imports
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest  # Linting
```

## 🚨 **Common Mistakes to Avoid**

1. **Ignoring errors** - Always handle `err`
2. **Not using `go fmt`** - Code must be formatted
3. **Exported everything** - Use lowercase for private stuff
4. **Complex functions** - Keep functions small and focused
5. **No error context** - Use `fmt.Errorf("context: %w", err)`

## 🎯 **Quick Checklist**

Before committing, ask yourself:
- ✅ Did I run `go fmt ./...`?
- ✅ Did I handle all errors?
- ✅ Are my function names clear?
- ✅ Did I use appropriate visibility (exported/unexported)?
- ✅ Does `go vet ./...` pass?
- ✅ Does the code build?

## 📚 **Learn More**

- [Effective Go](https://golang.org/doc/effective_go.html) - Official guide
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) - Common review feedback
- [Go by Example](https://gobyexample.com/) - Practical examples

---

**Remember: Go values simplicity and clarity over cleverness!** 🚀

*"Don't be clever, be clear"* - Go philosophy
