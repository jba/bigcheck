# bigcheck

This is Go analyzer that flags copying of large values.

Run it on some code and pass a -size flag:
```
bigcheck -size 200 program.go
```

It will report any statement where it can detect that a value of 200 bytes or
larger is being copied.



