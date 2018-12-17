package qcloud

func StringV(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
