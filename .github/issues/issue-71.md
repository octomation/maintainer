---
id: 71
database_id: 1315731752
node_id: I_kwDOE2M9Zc5ObHko
status: closed
title: "pkg: file: register encoders for specific formats"
labels: [scope: code, scope: test]
url: https://github.com/octomation/maintainer/issues/71
created_at: 2022-07-23T19:38:09Z
updated_at: 2022-07-25T19:29:56Z
---

# pkg: file: register encoders for specific formats

**Motivation:** it allows to simplify code like this

```go
	var data HeatMap
	format := strings.ToLower(filepath.Ext(file.Name()))
	switch format {
	case ".json":
		err := json.NewDecoder(file).Decode(&data)
		src.data = data
		return data, err
	case ".yml", ".yaml":
		err := yaml.NewDecoder(file).Decode(&data)
		src.data = data
		return data, err
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
```

```go
func pack(file afero.File, data any) error {
	format := strings.ToLower(filepath.Ext(file.Name()))

	switch format {
	case ".json":
		return json.NewEncoder(file).Encode(data)
	case ".yml", ".yaml":
		return yaml.NewEncoder(file).Encode(data)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}

func unpack(file afero.File, ptr any) error {
	format := strings.ToLower(filepath.Ext(file.Name()))

	switch format {
	case ".json":
		return json.NewDecoder(file).Decode(ptr)
	case ".yml", ".yaml":
		return yaml.NewDecoder(file).Decode(ptr)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}
}
```
