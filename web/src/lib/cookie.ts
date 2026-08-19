import Cookies from 'js-cookie';

const COOKIE_TOKEN_KEY = 'nano-kvm-token';

export function existToken() {
  const token = Cookies.get(COOKIE_TOKEN_KEY);
  return !!token;
}

export function getToken() {
  const token = Cookies.get(COOKIE_TOKEN_KEY);
  if (!token) return null;

  return token;
}

// setToken stores the session and reports whether the browser kept it.
//
// The answer is not always yes, and a caller that assumes it is produces a
// login screen that accepts the password and returns to itself forever.
//
// A cookie written over https carries Secure. It is then not sent over plain
// http, AND a page served over plain http may not overwrite or delete it: that
// is RFC 6265bis section 5.4, "Leave Secure Cookies Alone", which every current
// browser enforces. So after HTTPS is switched off, the old Secure cookie is
// invisible and immovable, every login writes a cookie the browser discards,
// and the application cannot clear it either. The operator has to clear cookies
// by hand, and nothing tells them so.
//
// The Tls switch deletes the token before it changes the scheme, which is the
// fix. This return value is the net under it, for a browser that arrives in
// that state some other way.
export function setToken(token: string): boolean {
  // sameSite strict keeps the token off cross-site requests, which is what
  // stops another page from driving the device through the API or a websocket.
  // secure is only set on https, otherwise the browser would drop the cookie
  // on devices served over plain http.
  Cookies.set(COOKIE_TOKEN_KEY, token, {
    expires: 30,
    sameSite: 'strict',
    secure: window.location.protocol === 'https:'
  });

  return Cookies.get(COOKIE_TOKEN_KEY) === token;
}

export function removeToken() {
  Cookies.remove(COOKIE_TOKEN_KEY);
}
