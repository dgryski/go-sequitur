package sequitur

import (
	"bytes"
	"fmt"
)

const testString = `
pease porridge hot,
pease porridge cold,
pease porridge in the pot,
nine days old.

some like it hot,
some like it cold,
some like it in the pot,
nine days old.

《施氏食狮史》
石室诗士施氏，嗜狮，誓食十狮。
氏时时适市视狮。
十时，适十狮适市。
是时，适施氏适市。
氏视是十狮，恃矢势，使是十狮逝世。
氏拾是十狮尸，适石室。
石室湿，氏使侍拭石室。
石室拭，氏始试食是十狮尸。
食时，始识是十狮，实十石狮尸。
试释是事。
` // https://www.duolingo.com/comment/25323797/Fun-Chinese-poem-The-Lion-Eating-Poet-in-the-Stone-Den

func ExamplePrettyPrintUTF8() {

	g := Parse([]byte(testString))

	var output bytes.Buffer
	if err := g.PrettyPrint(&output); err != nil {
		panic(err)
	}

	fmt.Println(output.String())

	// Output:
	// 0 -> 1 2 3 4 3 5 6 2 7 4 7 5 《 8 食 狮 史 》 \n 9 诗 士 8 ， 嗜 狮 ， 誓 食 十 10 氏 时 时 11 视 10 十 12 13 14 是 12 8 14 氏 视 15 恃 矢 势 ， 使 16 逝 世 17 氏 拾 18 19 20 湿 21 使 侍 拭 20 拭 21 始 试 食 18 17 食 时 ， 始 识 15 实 十 石 狮 尸 17 试 释 是 事 17
	// 1 -> \n p e a s 22 r r i d g 23
	// 2 -> h o t
	// 3 -> , 1
	// 4 -> c 24
	// 5 -> 25 _ t h 22 t 26 n 25 23 d a y s _ 24 . \n \n
	// 6 -> s o m 23 l i k 23 i t _
	// 7 -> 26 6
	// 8 -> 施 氏
	// 9 -> 石 室
	// 10 -> 狮 17
	// 11 -> 适 市
	// 12 -> 时 19
	// 13 -> 十 狮
	// 14 -> 11 17
	// 15 -> 16 ，
	// 16 -> 是 13
	// 17 -> 。 \n
	// 18 -> 16 尸
	// 19 -> ， 适
	// 20 -> 9 17 9
	// 21 -> ， 氏
	// 22 -> 23 p o
	// 23 -> e _
	// 24 -> o l d
	// 25 -> i n
	// 26 -> , \n
}
