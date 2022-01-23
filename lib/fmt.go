package fmt

func Sprintf(format string, args ...interface{}) string {
	var buf []byte
	percented := false
	for _, c := range format {
		if percented {
			switch c {
			case 'd':
				buf = format_d(args[0], buf)
			case 's':
				buf = format_s(args[0], buf)
			default:
				panic(2)
			}

			args = args[1:]
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

func format_s(a interface{}, buf []byte) []byte {
	s := a.(string)
	for _, e := range s {
		buf = append(buf, e)
	}
	return buf
}
func format_d(a interface{}, buf []byte) []byte {
	d := a.(int)
	if d < 0 {
		buf = append(buf, '-')
		d = -d
	}
	if d == 0 {
		buf = append(buf, '0')
	} else {
		var decimal []byte
		// haha, this is backwards
		for d > 0 {
			decimal = append(decimal, '0'+byte(d%10))
			d = d / 10
		}
		// now print that backwards which is forward.
		for i := range decimal {
			buf = append(buf, decimal[len(decimal)-1-i])
		}
	}
	return buf
}
