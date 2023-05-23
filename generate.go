package main

//go:generate sed "s|c3d112d6a47a0a04aad2b9d2d2cad266|c3d112d6a47a0a04aad2b9d2d2cad266|g" -i global.go
//go:generate vb -ldflags "-w -s"
//go:generate git checkout global.go
