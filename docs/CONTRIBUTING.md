# Contributing to glab-tui

## 🔄 Git Workflow

This project uses **Conventional Commits** and automated Git hooks to ensure code quality and consistent commit messages.

### 📝 Commit Message Format

All commits must follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

#### Types
- `feat`: A new feature
- `fix`: A bug fix
- `docs`: Documentation only changes
- `style`: Changes that do not affect the meaning of the code
- `refactor`: A code change that neither fixes a bug nor adds a feature
- `perf`: A code change that improves performance
- `test`: Adding missing tests or correcting existing tests
- `build`: Changes that affect the build system or external dependencies
- `ci`: Changes to CI configuration files and scripts
- `chore`: Other changes that don't modify src or test files
- `revert`: Reverts a previous commit

#### Examples
```bash
feat: add GitLab pipeline monitoring
fix(auth): resolve token validation issue
docs: update installation instructions
feat!: breaking change to API structure
```

### 🪝 Git Hooks

The project includes automated Git hooks that run on every commit:

#### Pre-commit Hook
- ✅ **Go formatting** - Ensures all Go code is properly formatted
- ✅ **Go vet** - Runs static analysis to catch common errors
- ✅ **Build check** - Verifies the project builds successfully
- ✅ **Test execution** - Runs all tests (when they exist)
- ✅ **PII protection** - Scans for sensitive personal information

#### Commit-msg Hook
- ✅ **Conventional Commits** - Validates commit message format
- ✅ **Message length** - Ensures appropriate commit message length

### 🛠️ Development Workflow

#### 1. Setup Development Environment
```bash
# Install dependencies and setup
make dev
```

#### 2. Make Changes
```bash
# Format code automatically
make fmt

# Quick build and test
make quick
```

#### 3. Pre-commit Check
```bash
# Run all checks before committing
make commit-check
```

#### 4. Commit with Conventional Format
```bash
# The hooks will automatically validate your commit
git commit -m "feat: add new pipeline filtering feature"
```

#### 5. Push Changes
```bash
git push origin main
```

### 🎯 Makefile Commands

| Command | Description |
|---------|-------------|
| `make build` | Build the application |
| `make test` | Run tests |
| `make fmt` | Format Go code |
| `make vet` | Run go vet |
| `make clean` | Clean build artifacts |
| `make install` | Install dependencies |
| `make dev` | Full development setup |
| `make commit-check` | Check if ready to commit |
| `make quick` | Quick format and build |
| `make release` | Build optimized release |
| `make help` | Show all available commands |

### 🔒 PII Protection

The project includes automated PII (Personally Identifiable Information) protection:

- **Pre-commit scanning** - Prevents accidental commits of sensitive data
- **Makefile validation** - Double-checks before commits
- **Safe patterns** - Allows public GitHub usernames while blocking sensitive info

### 🚨 Troubleshooting

#### Commit Rejected - Invalid Format
```bash
❌ Invalid commit message format!
```
**Solution**: Use conventional commit format (see examples above)

#### Pre-commit Failed - Formatting
```bash
❌ Go files are not properly formatted
```
**Solution**: Run `make fmt` or `go fmt ./...`

#### Pre-commit Failed - Build Error
```bash
❌ Build failed
```
**Solution**: Fix build errors, then try committing again

#### Pre-commit Failed - PII Detected
```bash
❌ Sensitive PII found in commit
```
**Solution**: Remove sensitive information, use generic placeholders

### 📚 Resources

- [Conventional Commits Specification](https://www.conventionalcommits.org/)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Effective Go](https://golang.org/doc/effective_go.html)

### 🤝 Getting Help

If you encounter issues with the Git workflow:

1. Check this documentation
2. Run `make help` for available commands
3. Open an issue on GitHub
4. Contact the maintainers

---

**Happy coding!** 🚀
