# GoStructMap

`GoStructMap` - библиотека, которая упрощает процесс отображения структур данных, подобных JSON (таких как `map[string]interface{}`), в структуры Go и наоборот, обеспечивая безопасность типов и обработку ошибок.

## Features
- Map JSON-like в Go структуры.
- Обработка базовых типов (`int`, `float`, `bool`, `string`, etc.).
- Поддержка slices, arrays and maps.
- Поддержка вложенных структур и указателей.

## Установка

```bash
go get github.com/bovinxx/gostructmap
```

## Использование
Пример

```go
package main

import (
	"fmt"
	"github.com/bovinxx/gostructmap"
)

type Person struct {
	ID       int
	Username string
	Active   bool
}

func main() {
	data := map[string]interface{}{
		"ID":       1,
		"Username": "john_doe",
		"Active":   true,
	}

	var decoder gostructmap.Decoder = gostructmap.NewDecoder()

	var person Person
	err := decoder.Decode(data, &person)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Printf("Mapped struct: %+v\n", person)
}
```
