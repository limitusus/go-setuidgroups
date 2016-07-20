# go-setuidgroups

golang implementation of setuidgroups

# Runtime environments

* works on OS X
* DOES NOT work on Linux, since `syscall.Setgid` was disabled on Linux from Go 1.4.
  * See: https://golang.org/doc/go1.4#minor_library_changes
