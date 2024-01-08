# type (monitor,printer,lis)
build:
	export CC=x86_64-w64-mingw32-gcc
	export CGO_ENABLED=0
	export GOARCH=amd64
	export GOOS=windows
	fyne package -os windows -icon tana.png --name "Tana App"

macbuild:
	export GOOS=darwin
	fyne package -os darwin -icon tana.png --name "Tana App"

