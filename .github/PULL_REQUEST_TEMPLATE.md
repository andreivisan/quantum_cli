## Description
<!-- Provide a brief summary of the changes. Explain the problem you're solving and why the solution works. -->

## Related Issues
<!-- Mention related issues or feature requests, if any. For example, "Closes #123". -->

## Type of Change
Please mark the relevant options:

- [ ] 🐛 Bug fix
- [ ] 🚀 New feature
- [ ] 📝 Documentation update
- [ ] 🔧 Refactoring
- [ ] ⚡️ Performance improvement
- [ ] ✅ Test addition
- [ ] 🔒 Security fix

## Checklist
Please ensure your code adheres to the following best practices before submitting the pull request:
- [ ] **Readability**: Minimized nested blocks and aligned the happy path to the left for better readability.
- [ ] **Variable Management**: Avoided variable shadowing and used descriptive, meaningful variable names.
- [ ] **Error Handling**:
  - Properly handled errors without relying on `panic` or ignoring errors silently.
  - Wrapped errors when propagating them for better context.
- [ ] **Concurrency**:
  - Verified that goroutines terminate appropriately to prevent leaks.
  - Ensured correct use of channels and mutexes to avoid race conditions.
- [ ] **Testing**:
  - Added unit tests or updated existing ones to cover the changes.
  - Enabled `-race` flag when running tests to detect race conditions.
  - Avoided dependencies on global variables for testability.
- [ ] **Documentation**:
  - Added or updated comments to ensure code is understandable.
  - Ensured public functions and packages include proper GoDoc comments.

## Linting and Formatting
- [ ] Ran `go fmt` and `go vet` to ensure the code is properly formatted and linted.
- [ ] Verified that there are no linting issues by running `golangci-lint` or a similar tool.

## API Changes (If Applicable)
- [ ] Ensured any changes to the API maintain backward compatibility or include proper documentation for breaking changes.
- [ ] Avoided interface pollution by keeping interfaces minimal and specific.

## Performance
- [ ] Verified that the code avoids unnecessary allocations and memory leaks.
- [ ] Ensured optimized string or slice operations, if applicable.

## Additional Notes
<!-- Add any additional information or context for the reviewers. -->
