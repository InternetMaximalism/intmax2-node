package http_response_modifier

import (
	"context"
	"fmt"
	"intmax2-node/configs"
	"intmax2-node/internal/pb/gateway/consts"
	"net/http"
	"strconv"
	"strings"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/prodadidb/go-email-validator/pkg/ev"
	"github.com/prodadidb/go-email-validator/pkg/ev/evmail"
	"github.com/prodadidb/go-validation"
	"google.golang.org/protobuf/proto"
)

type httpResponseModifier struct {
	ctx     context.Context
	w       http.ResponseWriter
	md      *runtime.ServerMetadata
	msg     proto.Message
	cfg     *configs.Config
	cookies *Cookies
}

type HttpModifier interface {
	Apply() error
}

func New(
	ctx context.Context,
	w http.ResponseWriter,
	md *runtime.ServerMetadata,
	msg proto.Message,
	cookies *Cookies,
) HttpModifier {
	return &httpResponseModifier{
		ctx:     ctx,
		w:       w,
		md:      md,
		msg:     msg,
		cookies: cookies,
		cfg:     ctx.Value(consts.AppConfigs).(*configs.Config),
	}
}

func (m *httpResponseModifier) autoRemove(header string) {
	const split = "-"
	parts := strings.Split(header, split)
	for key := range parts {
		if parts[key] == "" {
			continue
		}
		item := strings.ToUpper(parts[key][:1])
		l := len([]rune(parts[key]))
		if l > 1 {
			item += strings.ToLower(parts[key][1:l])
		}
		parts[key] = item
	}
	header = strings.Join(parts, split)
	delete(m.w.Header(), fmt.Sprintf(defaultMask, defaultValue, header))
}

func (m *httpResponseModifier) cookie(name, value string, maxAgeSecond int) {
	cookie := http.Cookie{
		Name:     name,
		Path:     "/",
		Value:    value,
		HttpOnly: true,
		MaxAge:   maxAgeSecond,
	}

	if m.cookies.Secure {
		cookie.Secure = true
		cookie.SameSite = http.SameSiteNoneMode
	}

	if m.cookies.Domain != "" {
		const mockEmailValidator = "mock@"
		if validation.Validate(m.cookies.Domain, validation.Required, validation.By(func(value interface{}) error {
			val, ok := value.(string)
			if !ok {
				return ErrValueInvalid
			}

			if v := ev.NewSyntaxValidator().Validate(ev.NewInput(evmail.FromString(mockEmailValidator + val))); !v.IsValid() {
				return ErrValueInvalid
			}

			return nil
		})) == nil {
			cookie.Domain = m.cookies.Domain

			if m.cookies.SameSiteStrictMode {
				cookie.SameSite = http.SameSiteStrictMode
			}
		}
	}

	http.SetCookie(m.w, &cookie)
}

func (m *httpResponseModifier) Apply() error {
	if err := m.outDelCookieAccessToken(); err != nil {
		return err
	}
	if err := m.outDelCookieWalletAddress(); err != nil {
		return err
	}
	if err := m.httpCode(); err != nil {
		return err
	}

	return nil
}

func (m *httpResponseModifier) outDelCookieAccessToken() (err error) {
	defer m.autoRemove(OutDelCookieAccessToken)
	h := m.md.HeaderMD.Get(OutDelCookieAccessToken)
	if len(h) == 0 {
		return nil
	}

	const int1Key = 1
	m.cookie(consts.AccessToken, h[0], int1Key)

	return nil
}

func (m *httpResponseModifier) outDelCookieWalletAddress() (err error) {
	defer m.autoRemove(OutDelCookieWalletAddress)
	h := m.md.HeaderMD.Get(OutDelCookieWalletAddress)
	if len(h) == 0 {
		return nil
	}

	const int1Key = 1
	m.cookie(consts.WalletAddress, h[0], int1Key)

	return nil
}

func (m *httpResponseModifier) httpCode() error {
	h := m.md.HeaderMD.Get(OutHTTPCode)
	if len(h) == 0 {
		return nil
	}

	code, err := strconv.Atoi(h[0])
	if err != nil {
		return err
	}

	delete(m.md.HeaderMD, OutHTTPCode)
	m.autoRemove(OutHTTPCode)
	m.w.WriteHeader(code)

	return nil
}
