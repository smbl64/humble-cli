# Finding the session key in Firefox

1. Go to the [Humble Bundle][hb] website
2. Make sure you are logged in
3. Press `F12` to open the Web Developer Tools
4. Go to the **Storage** tab in the Web Developer Tools
5. On the left pane, Select **Cookies** and then **https://www.humblebundle.com**
6. On the right pane, click on the row whose name starts with **`ory_session_`** (Humble's login moved to `auth.humblebundle.com`, so the old `_simpleauth_sess` cookie is gone)
7. Double-click on the **Value** column and copy the selected value
8. The cookie is HttpOnly, so `document.cookie` will not show it; copy it from the Developer Tools instead

Pass the cookie to the CLI as `name=value`, including the `ory_session_` prefix:

```sh
humble-cli auth 'ory_session_abc123=<value>'
```

For a visual guide on steps 4 to 7, [see this picture](firefox.jpg).

[hb]: https://www.humblebundle.com
