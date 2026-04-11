package service

func stringSliceFromAny(raw interface{}) []string {
	switch value := raw.(type) {
	case nil:
		return nil
	case []string:
		return append([]string(nil), value...)
	case []any:
		result := make([]string, 0, len(value))
		for _, item := range value {
			switch typed := item.(type) {
			case string:
				result = append(result, typed)
			case []byte:
				result = append(result, string(typed))
			}
		}
		return result
	default:
		return nil
	}
}
