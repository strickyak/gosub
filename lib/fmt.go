package fmt

func Sprintf(format string, args ...interface{}) string {
	var buf []byte
	percented := false
	for _, c := range format {
		if percented {
			switch c {
			case 'd':
				d := args[0].(int)
				args = args[1:]
				if d < 0 {
					buf = append(buf, '-')
					d = -d
				}
				if d == 0 {
					buf = append(buf, '0')
				} else {
					// haha, this is backwards
					for d > 0 {
						buf = append(buf, '0'+byte(d%10))
						d = d / 10
					}
				}
			case 's':
				s := args[0].(string)
				for _, e := range s {
					buf = append(buf, e)
				}
			default:
				panic(2)
			}

			percented = false
		} else {
			if c == '%' {
				percented = true
			} else {
				buf = append(buf, byte(c))
			}
		}
	}
	return string(buf)
}
