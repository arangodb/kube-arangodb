//
// DISCLAIMER
//
// Copyright 2017 ArangoDB GmbH, Cologne, Germany
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// Copyright holder is ArangoDB GmbH, Cologne, Germany
//
// Author Ewout Prangsma
//

package test

import (
	"bytes"
	"testing"

	velocypack "github.com/arangodb/go-velocypack"
)

func TestDumperNull(t *testing.T) {
	s := velocypack.Slice([]byte{0x18})
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("null", buf.String(), t)
}

func TestDumperFalse(t *testing.T) {
	s := velocypack.Slice([]byte{0x19})
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("false", buf.String(), t)
}

func TestDumperTrue(t *testing.T) {
	s := velocypack.Slice([]byte{0x1a})
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("true", buf.String(), t)
}

func TestDumperStringSimple(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("foobar")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ(`"foobar"`, buf.String(), t)
}

func TestDumperStringSpecialChars(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\"fo\r \n \\to''\\ \\bar\"")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("\"\\\"fo\\r \\n \\\\to''\\\\ \\\\bar\\\"\"", buf.String(), t)
}

func TestDumperStringControlChars(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\x00\x01\x02 baz \x03")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("\"\\u0000\\u0001\\u0002 baz \\u0003\"", buf.String(), t)
}

func TestDumperStringUTF8(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("mötör")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("\"mötör\"", buf.String(), t)
}

func TestDumperStringUTF8Escaped(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("mötör")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, &velocypack.DumperOptions{EscapeUnicode: true})
	d.Append(s)
	ASSERT_EQ("\"m\\u00F6t\\u00F6r\"", buf.String(), t)
}

func TestDumperStringTwoByteUTF8(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\xc2\xa2")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("\"\xc2\xa2\"", buf.String(), t)
}

func TestDumperStringTwoByteUTF8Escaped(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\xc2\xa2")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, &velocypack.DumperOptions{EscapeUnicode: true})
	d.Append(s)
	ASSERT_EQ("\"\\u00A2\"", buf.String(), t)
}

func TestDumperStringThreeByteUTF8(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\xe2\x82\xac")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("\"\xe2\x82\xac\"", buf.String(), t)
}

func TestDumperStringThreeByteUTF8Escaped(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\xe2\x82\xac")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, &velocypack.DumperOptions{EscapeUnicode: true})
	d.Append(s)
	ASSERT_EQ("\"\\u20AC\"", buf.String(), t)
}

func TestDumperStringFourByteUTF8(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\xf0\xa4\xad\xa2")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, nil)
	d.Append(s)
	ASSERT_EQ("\"\xf0\xa4\xad\xa2\"", buf.String(), t)
}

func TestDumperStringFourByteUTF8Escaped(t *testing.T) {
	b := velocypack.Builder{}
	must(b.AddValue(velocypack.NewStringValue("\xf0\xa4\xad\xa2")))

	s := mustSlice(b.Slice())
	buf := &bytes.Buffer{}
	d := velocypack.NewDumper(buf, &velocypack.DumperOptions{EscapeUnicode: true})
	d.Append(s)
	ASSERT_EQ("\"\\uD852\\uDF62\"", buf.String(), t)
}

func TestDumperStringMultiBytes(t *testing.T) {
	tests := []string{
		"Lorem ipsum dolor sit amet, te enim mandamus consequat ius, cu eos timeam bonorum, in nec eruditi tibique. At nec malorum saperet vivendo. Qui delectus moderatius in. Vivendo expetendis ullamcorper ut mei.",
		"Мёнём пауло пытынтёюм ад ыам. Но эрож рыпудяары вим, пожтэа эюрйпйдяч ентырпрытаряш ад хёз. Мыа дектаж дёжкэрэ котёдиэквюэ ан. Ведят брутэ мэдиокретатым йн прё",
		"Μει ει παρτεμ μολλις δελισατα, σιφιβυς σονσυλατυ ραθιονιβυς συ φις, φερι μυνερε μεα ετ. Ειρμωδ απεριρι δισενθιετ εα υσυ, κυο θωτα φευγαιθ δισενθιετ νο",
		"供覧必責同界要努新少時購止上際英連動信。同売宗島載団報音改浅研壊趣全。並嗅整日放横旅関書文転方。天名他賞川日拠隊散境行尚島自模交最葉駒到",
		"舞ばい支小ぜ館応ヌエマ得6備ルあ煮社義ゃフおづ報載通ナチセ東帯あスフず案務革た証急をのだ毎点十はぞじド。1芸キテ成新53験モワサセ断団ニカ働給相づらべさ境著ラさ映権護ミオヲ但半モ橋同タ価法ナカネ仙説時オコワ気社オ",
		"أي جنوب بداية السبب بلا. تمهيد التكاليف العمليات إذ دول, عن كلّ أراضي اعتداء, بال الأوروبي الإقتصادية و. دخول تحرّكت بـ حين. أي شاسعة لليابان استطاعوا مكن. الأخذ الصينية والنرويج هو أخذ.",
		"זכר דפים בדפים מה, צילום מדינות היא או, ארץ צרפתית העברית אירועים ב. שונה קולנוע מתן אם, את אחד הארץ ציור וכמקובל. ויש העיר שימושי מדויקים בה, היא ויקי ברוכים תאולוגיה או. את זכר קהילה חבריכם ליצירתה, ערכים התפתחות חפש גם.",
	}
	for _, test := range tests {
		b := velocypack.Builder{}
		must(b.AddValue(velocypack.NewStringValue(test)))

		s := mustSlice(b.Slice())
		buf := &bytes.Buffer{}
		d := velocypack.NewDumper(buf, nil)
		d.Append(s)
		expected := "\"" + test + "\""
		ASSERT_EQ(expected, buf.String(), t)
	}
}

func TestDumperStringMultiBytesEscaped(t *testing.T) {
	tests := map[string]string{
		"Мёнём пауло пытынтёюм ад ыам. Но эрож рыпудяары вим, пожтэа эюрйпйдяч ентырпрытаряш ад хёз. Мыа дектаж дёжкэрэ котёдиэквюэ ан. Ведят брутэ мэдиокретатым йн прё":                                                                "\\u041C\\u0451\\u043D\\u0451\\u043C \\u043F\\u0430\\u0443\\u043B\\u043E \\u043F\\u044B\\u0442\\u044B\\u043D\\u0442\\u0451\\u044E\\u043C \\u0430\\u0434 \\u044B\\u0430\\u043C. \\u041D\\u043E \\u044D\\u0440\\u043E\\u0436 \\u0440\\u044B\\u043F\\u0443\\u0434\\u044F\\u0430\\u0440\\u044B \\u0432\\u0438\\u043C, \\u043F\\u043E\\u0436\\u0442\\u044D\\u0430 \\u044D\\u044E\\u0440\\u0439\\u043F\\u0439\\u0434\\u044F\\u0447 \\u0435\\u043D\\u0442\\u044B\\u0440\\u043F\\u0440\\u044B\\u0442\\u0430\\u0440\\u044F\\u0448 \\u0430\\u0434 \\u0445\\u0451\\u0437. \\u041C\\u044B\\u0430 \\u0434\\u0435\\u043A\\u0442\\u0430\\u0436 \\u0434\\u0451\\u0436\\u043A\\u044D\\u0440\\u044D \\u043A\\u043E\\u0442\\u0451\\u0434\\u0438\\u044D\\u043A\\u0432\\u044E\\u044D \\u0430\\u043D. \\u0412\\u0435\\u0434\\u044F\\u0442 \\u0431\\u0440\\u0443\\u0442\\u044D \\u043C\\u044D\\u0434\\u0438\\u043E\\u043A\\u0440\\u0435\\u0442\\u0430\\u0442\\u044B\\u043C \\u0439\\u043D \\u043F\\u0440\\u0451",
		"Μει ει παρτεμ μολλις δελισατα, σιφιβυς σονσυλατυ ραθιονιβυς συ φις, φερι μυνερε μεα ετ. Ειρμωδ απεριρι δισενθιετ εα υσυ, κυο θωτα φευγαιθ δισενθιετ νο":                                                                         "\\u039C\\u03B5\\u03B9 \\u03B5\\u03B9 \\u03C0\\u03B1\\u03C1\\u03C4\\u03B5\\u03BC \\u03BC\\u03BF\\u03BB\\u03BB\\u03B9\\u03C2 \\u03B4\\u03B5\\u03BB\\u03B9\\u03C3\\u03B1\\u03C4\\u03B1, \\u03C3\\u03B9\\u03C6\\u03B9\\u03B2\\u03C5\\u03C2 \\u03C3\\u03BF\\u03BD\\u03C3\\u03C5\\u03BB\\u03B1\\u03C4\\u03C5 \\u03C1\\u03B1\\u03B8\\u03B9\\u03BF\\u03BD\\u03B9\\u03B2\\u03C5\\u03C2 \\u03C3\\u03C5 \\u03C6\\u03B9\\u03C2, \\u03C6\\u03B5\\u03C1\\u03B9 \\u03BC\\u03C5\\u03BD\\u03B5\\u03C1\\u03B5 \\u03BC\\u03B5\\u03B1 \\u03B5\\u03C4. \\u0395\\u03B9\\u03C1\\u03BC\\u03C9\\u03B4 \\u03B1\\u03C0\\u03B5\\u03C1\\u03B9\\u03C1\\u03B9 \\u03B4\\u03B9\\u03C3\\u03B5\\u03BD\\u03B8\\u03B9\\u03B5\\u03C4 \\u03B5\\u03B1 \\u03C5\\u03C3\\u03C5, \\u03BA\\u03C5\\u03BF \\u03B8\\u03C9\\u03C4\\u03B1 \\u03C6\\u03B5\\u03C5\\u03B3\\u03B1\\u03B9\\u03B8 \\u03B4\\u03B9\\u03C3\\u03B5\\u03BD\\u03B8\\u03B9\\u03B5\\u03C4 \\u03BD\\u03BF",
		"供覧必責同界要努新少時購止上際英連動信。同売宗島載団報音改浅研壊趣全。並嗅整日放横旅関書文転方。天名他賞川日拠隊散境行尚島自模交最葉駒到":                                                                                                                                                           "\\u4F9B\\u89A7\\u5FC5\\u8CAC\\u540C\\u754C\\u8981\\u52AA\\u65B0\\u5C11\\u6642\\u8CFC\\u6B62\\u4E0A\\u969B\\u82F1\\u9023\\u52D5\\u4FE1\\u3002\\u540C\\u58F2\\u5B97\\u5CF6\\u8F09\\u56E3\\u5831\\u97F3\\u6539\\u6D45\\u7814\\u58CA\\u8DA3\\u5168\\u3002\\u4E26\\u55C5\\u6574\\u65E5\\u653E\\u6A2A\\u65C5\\u95A2\\u66F8\\u6587\\u8EE2\\u65B9\\u3002\\u5929\\u540D\\u4ED6\\u8CDE\\u5DDD\\u65E5\\u62E0\\u968A\\u6563\\u5883\\u884C\\u5C1A\\u5CF6\\u81EA\\u6A21\\u4EA4\\u6700\\u8449\\u99D2\\u5230",
		"舞ばい支小ぜ館応ヌエマ得6備ルあ煮社義ゃフおづ報載通ナチセ東帯あスフず案務革た証急をのだ毎点十はぞじド。1芸キテ成新53験モワサセ断団ニカ働給相づらべさ境著ラさ映権護ミオヲ但半モ橋同タ価法ナカネ仙説時オコワ気社オ":                                                                                                                     "\\u821E\\u3070\\u3044\\u652F\\u5C0F\\u305C\\u9928\\u5FDC\\u30CC\\u30A8\\u30DE\\u5F976\\u5099\\u30EB\\u3042\\u716E\\u793E\\u7FA9\\u3083\\u30D5\\u304A\\u3065\\u5831\\u8F09\\u901A\\u30CA\\u30C1\\u30BB\\u6771\\u5E2F\\u3042\\u30B9\\u30D5\\u305A\\u6848\\u52D9\\u9769\\u305F\\u8A3C\\u6025\\u3092\\u306E\\u3060\\u6BCE\\u70B9\\u5341\\u306F\\u305E\\u3058\\u30C9\\u30021\\u82B8\\u30AD\\u30C6\\u6210\\u65B053\\u9A13\\u30E2\\u30EF\\u30B5\\u30BB\\u65AD\\u56E3\\u30CB\\u30AB\\u50CD\\u7D66\\u76F8\\u3065\\u3089\\u3079\\u3055\\u5883\\u8457\\u30E9\\u3055\\u6620\\u6A29\\u8B77\\u30DF\\u30AA\\u30F2\\u4F46\\u534A\\u30E2\\u6A4B\\u540C\\u30BF\\u4FA1\\u6CD5\\u30CA\\u30AB\\u30CD\\u4ED9\\u8AAC\\u6642\\u30AA\\u30B3\\u30EF\\u6C17\\u793E\\u30AA",
		"أي جنوب بداية السبب بلا. تمهيد التكاليف العمليات إذ دول, عن كلّ أراضي اعتداء, بال الأوروبي الإقتصادية و. دخول تحرّكت بـ حين. أي شاسعة لليابان استطاعوا مكن. الأخذ الصينية والنرويج هو أخذ.":                                     "\\u0623\\u064A \\u062C\\u0646\\u0648\\u0628 \\u0628\\u062F\\u0627\\u064A\\u0629 \\u0627\\u0644\\u0633\\u0628\\u0628 \\u0628\\u0644\\u0627. \\u062A\\u0645\\u0647\\u064A\\u062F \\u0627\\u0644\\u062A\\u0643\\u0627\\u0644\\u064A\\u0641 \\u0627\\u0644\\u0639\\u0645\\u0644\\u064A\\u0627\\u062A \\u0625\\u0630 \\u062F\\u0648\\u0644, \\u0639\\u0646 \\u0643\\u0644\\u0651 \\u0623\\u0631\\u0627\\u0636\\u064A \\u0627\\u0639\\u062A\\u062F\\u0627\\u0621, \\u0628\\u0627\\u0644 \\u0627\\u0644\\u0623\\u0648\\u0631\\u0648\\u0628\\u064A \\u0627\\u0644\\u0625\\u0642\\u062A\\u0635\\u0627\\u062F\\u064A\\u0629 \\u0648. \\u062F\\u062E\\u0648\\u0644 \\u062A\\u062D\\u0631\\u0651\\u0643\\u062A \\u0628\\u0640 \\u062D\\u064A\\u0646. \\u0623\\u064A \\u0634\\u0627\\u0633\\u0639\\u0629 \\u0644\\u0644\\u064A\\u0627\\u0628\\u0627\\u0646 \\u0627\\u0633\\u062A\\u0637\\u0627\\u0639\\u0648\\u0627 \\u0645\\u0643\\u0646. \\u0627\\u0644\\u0623\\u062E\\u0630 \\u0627\\u0644\\u0635\\u064A\\u0646\\u064A\\u0629 \\u0648\\u0627\\u0644\\u0646\\u0631\\u0648\\u064A\\u062C \\u0647\\u0648 \\u0623\\u062E\\u0630.",
		"זכר דפים בדפים מה, צילום מדינות היא או, ארץ צרפתית העברית אירועים ב. שונה קולנוע מתן אם, את אחד הארץ ציור וכמקובל. ויש העיר שימושי מדויקים בה, היא ויקי ברוכים תאולוגיה או. את זכר קהילה חבריכם ליצירתה, ערכים התפתחות חפש גם.": "\\u05D6\\u05DB\\u05E8 \\u05D3\\u05E4\\u05D9\\u05DD \\u05D1\\u05D3\\u05E4\\u05D9\\u05DD \\u05DE\\u05D4, \\u05E6\\u05D9\\u05DC\\u05D5\\u05DD \\u05DE\\u05D3\\u05D9\\u05E0\\u05D5\\u05EA \\u05D4\\u05D9\\u05D0 \\u05D0\\u05D5, \\u05D0\\u05E8\\u05E5 \\u05E6\\u05E8\\u05E4\\u05EA\\u05D9\\u05EA \\u05D4\\u05E2\\u05D1\\u05E8\\u05D9\\u05EA \\u05D0\\u05D9\\u05E8\\u05D5\\u05E2\\u05D9\\u05DD \\u05D1. \\u05E9\\u05D5\\u05E0\\u05D4 \\u05E7\\u05D5\\u05DC\\u05E0\\u05D5\\u05E2 \\u05DE\\u05EA\\u05DF \\u05D0\\u05DD, \\u05D0\\u05EA \\u05D0\\u05D7\\u05D3 \\u05D4\\u05D0\\u05E8\\u05E5 \\u05E6\\u05D9\\u05D5\\u05E8 \\u05D5\\u05DB\\u05DE\\u05E7\\u05D5\\u05D1\\u05DC. \\u05D5\\u05D9\\u05E9 \\u05D4\\u05E2\\u05D9\\u05E8 \\u05E9\\u05D9\\u05DE\\u05D5\\u05E9\\u05D9 \\u05DE\\u05D3\\u05D5\\u05D9\\u05E7\\u05D9\\u05DD \\u05D1\\u05D4, \\u05D4\\u05D9\\u05D0 \\u05D5\\u05D9\\u05E7\\u05D9 \\u05D1\\u05E8\\u05D5\\u05DB\\u05D9\\u05DD \\u05EA\\u05D0\\u05D5\\u05DC\\u05D5\\u05D2\\u05D9\\u05D4 \\u05D0\\u05D5. \\u05D0\\u05EA \\u05D6\\u05DB\\u05E8 \\u05E7\\u05D4\\u05D9\\u05DC\\u05D4 \\u05D7\\u05D1\\u05E8\\u05D9\\u05DB\\u05DD \\u05DC\\u05D9\\u05E6\\u05D9\\u05E8\\u05EA\\u05D4, \\u05E2\\u05E8\\u05DB\\u05D9\\u05DD \\u05D4\\u05EA\\u05E4\\u05EA\\u05D7\\u05D5\\u05EA \\u05D7\\u05E4\\u05E9 \\u05D2\\u05DD.",
	}
	for test, expected := range tests {
		b := velocypack.Builder{}
		must(b.AddValue(velocypack.NewStringValue(test)))

		s := mustSlice(b.Slice())
		buf := &bytes.Buffer{}
		d := velocypack.NewDumper(buf, &velocypack.DumperOptions{EscapeUnicode: true})
		d.Append(s)
		expected = "\"" + expected + "\""
		ASSERT_EQ(expected, buf.String(), t)
	}
}

func TestDumperDouble(t *testing.T) {
	tests := []struct {
		Value    float64
		Expected string
	}{
		{0.0, "0"},
		{123456.67, "123456.67"},
		{-123456.67, "-123456.67"},
		{-0.000442, "-0.000442"},
		{0.1, "0.1"},
		{2.41e-109, "2.41e-109"},
		{-3.423e+78, "-3.423e+78"},
		{3.423e+123, "3.423e+123"},
		{3.4239493e+104, "3.4239493e+104"},
	}
	for _, test := range tests {
		b := velocypack.Builder{}
		must(b.AddValue(velocypack.NewDoubleValue(test.Value)))

		s := mustSlice(b.Slice())
		buf := &bytes.Buffer{}
		d := velocypack.NewDumper(buf, nil)
		d.Append(s)
		ASSERT_EQ(test.Expected, buf.String(), t)
	}
}

func TestDumperInt(t *testing.T) {
	tests := []struct {
		Value    int64
		Expected string
	}{
		{0, "0"},
		{123456789, "123456789"},
		{-123456789, "-123456789"},
	}
	for _, test := range tests {
		b := velocypack.Builder{}
		must(b.AddValue(velocypack.NewIntValue(test.Value)))

		s := mustSlice(b.Slice())
		buf := &bytes.Buffer{}
		d := velocypack.NewDumper(buf, nil)
		d.Append(s)
		ASSERT_EQ(test.Expected, buf.String(), t)
	}
}

func TestDumperUInt(t *testing.T) {
	tests := []struct {
		Value    uint64
		Expected string
	}{
		{0, "0"},
		{5, "5"},
		{123456789, "123456789"},
	}
	for _, test := range tests {
		b := velocypack.Builder{}
		must(b.AddValue(velocypack.NewUIntValue(test.Value)))

		s := mustSlice(b.Slice())
		buf := &bytes.Buffer{}
		d := velocypack.NewDumper(buf, nil)
		d.Append(s)
		ASSERT_EQ(test.Expected, buf.String(), t)
	}
}

func TestDumperBinary(t *testing.T) {
	tests := []struct {
		Value    []byte
		Expected string
	}{
		{[]byte{1, 2, 3, 4}, "null"}, // Binary data is not supported by the Dumper
	}
	for _, test := range tests {
		b := velocypack.Builder{}
		must(b.AddValue(velocypack.NewBinaryValue(test.Value)))

		s := mustSlice(b.Slice())
		buf := &bytes.Buffer{}
		d := velocypack.NewDumper(buf, nil)
		d.Append(s)
		ASSERT_EQ(test.Expected, buf.String(), t)
	}
}
