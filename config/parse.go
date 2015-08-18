package config

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

const (
	LF   = '\n'
	CR   = '\r'
	CRLF = "\r\n"
	EOD  = ';'

	DONE        = 0
	BLOCK_START = 1
	PARSE_OK    = 2
	BLOCK_DONE  = 3
	IGNORE      = 4
	PARSE_ERR   = -1
)

func (c *Config) Parse(file string) error {
	var (
		err  error
		line []byte
		rc   int
	)

	if file != "" {
		prev := c.file
		prevBuf := c.buf
		c.file, err = os.Open(file)
		if err != nil {
			return err
		}

		c.buf = bufio.NewReader(c.file)

		defer func() {
			c.file.Close()
			c.file = prev
			c.buf = prevBuf
		}()
	}

	for {

		rc, err = c.readToken(line)
		if err != nil {
			return err
		}

		if rc == BLOCK_DONE {
			break
		}

		if rc == DONE {
			break
		}

		if len(c.args) > 0 {

			err = c.handler(rc)
			if err != nil {
				return err
			}
		}

	}

	return nil
}

func (c *Config) handler(last int) error {
	directive := c.args[0]

	for i := 0; i < len(c.modules); i++ {
		cmd := c.modules[i].Commands
		if len(cmd) == 0 {
			continue
		}

		for i := 0; i < len(cmd); i++ {
			if cmd[i].Name != directive {
				continue
			}

			if (cmd[i].T&BLOCK == 0) && last != PARSE_OK {
				return fmt.Errorf("directive \"%s\" is not terminated by \";\"", directive)
			}

			if (cmd[i].T&BLOCK > 0) && last != BLOCK_START {
				return fmt.Errorf("directive \"%s\" has no opening \"{\"", directive)
			}

			if cmd[i].T&ANY == 0 {

				if cmd[i].T&FLAG > 0 {

					if len(c.args) != 2 {
						goto invalid
					} else if cmd[i].T&MORE2 > 0 {

						if len(c.args) < 3 {
							goto invalid
						}
					}
				}
			}

			if err := cmd[i].Handler(c); err != nil {
				return err
			} else {
				return nil
			}
		}
	}

	return fmt.Errorf("invalid directive \"%s\"", directive)

invalid:
	return fmt.Errorf("invalid number of arguments in \"%s\" directive", directive)
}

func (c *Config) readToken(line []byte) (int, error) {
	var err error
	var ch byte
	quoted := false
	d_quoted := false
	s_quoted := false
	found := false
	need_space := false
	last_space := false
	variable := false
	c.args = []string{}

one:
	for {

		start := 0
		line, err = c.buf.ReadSlice(byte(LF))

		if err != nil && err != io.EOF {
			return PARSE_ERR, err
		}

		last_space = true

	two:
		for i := 0; i < len(line); i++ {

			ch = line[i]

			if ch == '#' {
				goto one
			}

			if quoted {
				quoted = false
				continue
			}

			if need_space {
				if ch == ' ' || ch == '\t' || ch == byte(CR) || ch == byte(LF) {
					last_space = true
					need_space = false
					continue
				}

				if ch == ';' {
					return PARSE_OK, nil
				}

				if ch == '{' {
					return BLOCK_START, nil
				}

				if ch == ')' {
					last_space = true
					need_space = false
				} else {
					return PARSE_ERR, fmt.Errorf("unexpected \"%c\"", ch)
				}
			}

			if last_space {

				if i != 0 {
					start++
				}

				if ch == ' ' || ch == '\t' || ch == byte(CR) || ch == byte(LF) {
					continue
				}

				switch ch {
				case ';':
				case '{':
					if ch == '{' {
						return BLOCK_START, nil
					}

					return PARSE_OK, nil
				case '}':
					return BLOCK_DONE, nil
				case '\\':
					quoted = true
					last_space = false
					start++
					continue two
				case '"':
					d_quoted = true
					last_space = false
					start++
					continue two
				case '\'':
					s_quoted = true
					last_space = false
					start++
					continue two
				default:
					last_space = false
				}
			} else {
				if ch == '{' && variable {
					continue
				}

				variable = false

				if ch == '\\' {
					quoted = true
					continue
				}

				if ch == '$' {
					variable = true
					continue
				}

				if d_quoted {
					if ch == '"' {
						d_quoted = false
						need_space = true
						found = true
					}
				} else if s_quoted {
					if ch == '\'' {
						s_quoted = false
						need_space = true
						found = true
					}
				} else if ch == ' ' || ch == '\t' || ch == byte(CR) ||
					ch == byte(LF) || ch == ';' || ch == '{' {
					last_space = true
					found = true
				}

				if found {
					word := []byte{}
					// fmt.Println(i, string(line[start:]))
				three:
					for ; start < i; start++ {

						if line[start] == '\\' {
							switch line[start+1] {
							case '"':
							case '\'':
							case '\\':
								start += 2
							case 't':
								word = append(word, '\t')
								start += 2
								goto three
							case 'r':
								word = append(word, '\r')
								start += 2
								goto three
							case 'n':
								word = append(word, '\n')
								start += 2
								goto three
							}
						}

						word = append(word, line[start])
					}

					if len(word) > 0 {
						c.args = append(c.args, string(word))
					}

					found = false

					if ch == ';' {
						return PARSE_OK, nil
					}

					if ch == '{' {
						return BLOCK_START, nil
					}
				}
			}
		}

		if err == io.EOF {

			if len(c.args) > 0 && !last_space {
				return PARSE_ERR, fmt.Errorf("unexpected end of file, expecting \";\" or \"}\"")
			}

			return DONE, nil
		}
	}

	return PARSE_OK, nil
}
