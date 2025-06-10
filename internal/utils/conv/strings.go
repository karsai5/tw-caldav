package conv

func SafeStringPtr(s *string) string {
	if s != nil {
		return *s
	}
	return ""
}
