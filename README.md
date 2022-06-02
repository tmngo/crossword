```bash
cd client/
yarn dev
```

```bash
cd server/
go run main.go
```

```
msdf-atlas-gen.exe -font AtkinsonHyperlegible-Regular.ttf ^
  -charset charset.txt ^
  -format png ^
  -imageout msdf-32-2.png ^
  -json msdf-32-2.json ^
  -size 32 ^
  -pxrange 2
```