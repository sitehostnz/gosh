### Self review

You should always review your own PR first.

For content changes, make sure that you:
- [ ] Confirm that the changes meet the user experience and goals outlined in the content design plan (if there is one).
- [ ] Review the content for technical accuracy.
- [ ] Run the following commands to make sure your code is clean, tested and adherence to the [style guide](https://github.com/sitehostnz/go-style-guide/blob/master/style.md).
```bash
# Making sure the repo is clean (no changes) 
make -B dirty

# Run tests
make vet

# Run linting
make lint
```
- [ ] If there are any failing checks in your PR, troubleshoot them until they're all passing.