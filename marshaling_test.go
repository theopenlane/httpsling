package httpsling

import (
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJSONMarshalerMarshal(t *testing.T) {
	m := JSONMarshaler{}

	v := map[string]interface{}{"color": "red"}

	expected, err := json.Marshal(v)
	require.NoError(t, err)

	expectedIndented, err := json.MarshalIndent(v, "", "  ")
	require.NoError(t, err)

	data, contentType, err := m.Marshal(v)
	require.NoError(t, err)
	require.Equal(t, "application/json;charset=utf-8", contentType)
	require.Equal(t, expected, data)

	m.Indent = true
	data, _, err = m.Marshal(v)
	require.NoError(t, err)
	require.Equal(t, expectedIndented, data)
}

func TestJSONMarshalerUnmarshal(t *testing.T) {
	m := JSONMarshaler{}
	d := []byte(`{"color":"red"}`)

	var v interface{}

	err := m.Unmarshal(d, "", &v)
	require.NoError(t, err)

	require.Equal(t, map[string]interface{}{"color": "red"}, v)
}

type testModel struct {
	Color string `xml:"color" json:"color" url:"color"`
	Count int    `xml:"count" json:"count" url:"count"`
}

func TestXMLMarshalerMarshal(t *testing.T) {
	m := XMLMarshaler{}

	b, ct, err := m.Marshal(testModel{"red", 30})
	require.NoError(t, err)

	assert.Equal(t, "application/xml;charset=utf-8", ct)

	assert.Equal(t, `<testModel><color>red</color><count>30</count></testModel>`, string(b))

	m.Indent = true
	b, _, err = m.Marshal(testModel{"red", 30})
	require.NoError(t, err)

	assert.Equal(t, `<testModel>
  <color>red</color>
  <count>30</count>
</testModel>`, string(b))
}

func TestXMLMarshalerUnmarshal(t *testing.T) {
	m := XMLMarshaler{}

	data := []byte(`<testModel><color>red</color><count>30</count></testModel>`)

	var v testModel

	err := m.Unmarshal(data, "", &v)
	require.NoError(t, err)

	assert.Equal(t, testModel{"red", 30}, v)
}

func TestContentTypeUnmarshalerUnmarshal(t *testing.T) {
	m := NewContentTypeUnmarshaler()
	m.Unmarshalers["another/thing"] = &JSONMarshaler{}

	cases := []struct {
		input       string
		contentType string
	}{
		{
			input:       `<testModel><color>red</color><count>30</count></testModel>`,
			contentType: `application/xml`,
		},
		{
			input:       `{"color":"red","count":30}`,
			contentType: `application/json`,
		},
		{
			input:       `{"color":"red","count":30}`,
			contentType: `application/tree.subtype+json`,
		},
		{
			input:       `{"color":"red","count":30}`,
			contentType: `another/thing`,
		},
	}
	for _, c := range cases {
		t.Run(c.contentType, func(t *testing.T) {
			var v testModel

			err := m.Unmarshal([]byte(c.input), c.contentType, &v)
			require.NoError(t, err)

			assert.Equal(t, testModel{"red", 30}, v)
		})
	}

	t.Run("unknown", func(t *testing.T) {
		err := m.Unmarshal([]byte(`{"color":"red","count":30}`), "application/unknown", &testModel{})
		require.Error(t, err)
	})

	t.Run("invalid media type", func(t *testing.T) {
		err := m.Unmarshal([]byte(`{"color":"red","count":30}`), "application|json", &testModel{})
		require.Error(t, err)
	})
}

func TestContentTypeUnmarshalerApply(t *testing.T) {
	r := MustNew()
	r.Marshaler = nil

	m := NewContentTypeUnmarshaler()
	r.MustApply(m)

	assert.Equal(t, m, r.Unmarshaler)
}

func TestFormMarshalerMarshal(t *testing.T) {
	testCases := []struct {
		input  interface{}
		output string
	}{
		{
			input:  testModel{"red", 30},
			output: "color=red&count=30",
		},
		{
			input:  map[string][]string{"color": {"green", "red"}, "count": {"40"}},
			output: "color=green&color=red&count=40",
		},
		{
			input:  url.Values{"color": {"green", "red"}, "count": {"40"}},
			output: "color=green&color=red&count=40",
		},
		{
			input:  map[string]string{"color": "green", "count": "40"},
			output: "color=green&count=40",
		},
	}
	for _, testCase := range testCases {
		m := FormMarshaler{}

		data, contentType, err := m.Marshal(testCase.input)
		require.NoError(t, err)

		assert.Equal(t, "application/x-www-form-urlencoded", contentType)
		assert.Equal(t, testCase.output, string(data))
	}
}

func TestMarshalFuncApply(t *testing.T) {
	var mf MarshalFunc = func(_ interface{}) (bytes []byte, s string, e error) {
		return nil, "red", nil
	}

	_, s, _ := MustNew(mf).Marshaler.Marshal(nil)
	assert.Equal(t, "red", s)
}

func ExampleFormMarshaler() {
	req, _ := Request(&FormMarshaler{}, Body(url.Values{"color": []string{"red"}}))

	b, _ := io.ReadAll(req.Body)

	fmt.Println(string(b))
	fmt.Println(req.Header.Get(HeaderContentType))

	// Output:
	// color=red
	// application/x-www-form-urlencoded
}

func ExampleJSONMarshaler() {
	req, _ := Request(&JSONMarshaler{Indent: false}, Body(map[string]interface{}{"color": "red"}))

	b, _ := io.ReadAll(req.Body)

	fmt.Println(string(b))
	fmt.Println(req.Header.Get(HeaderContentType))

	// Output:
	// {"color":"red"}
	// application/json;charset=utf-8
}

func ExampleXMLMarshaler() {
	type Resource struct {
		Color string
	}

	req, _ := Request(&XMLMarshaler{Indent: false}, Body(Resource{Color: "red"}))

	b, _ := io.ReadAll(req.Body)

	fmt.Println(string(b))
	fmt.Println(req.Header.Get(HeaderContentType))

	// Output:
	// <Resource><Color>red</Color></Resource>
	// application/xml;charset=utf-8
}
