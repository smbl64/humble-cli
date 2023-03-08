# Finding the session key in Safari

First on all, if you don't see the **Develop** menu Safari:

1. Press `Cmd + ,` to open the **Preferences**
2. Go to the **Advanced** tab
3. Check "Show Develop menu in menu bar"


Now to find the session key:


1. Go to the [Humble Bundle][hb] website
2. Make sure you are logged in
3. Press `Cmd + Shift + I` to open the Web Inspector (or from the menu bar select Develop > Show Web Inspector)
4. Select the **Storage** tab in the Web Inspector
5. On the left pane, select **Cookies**
6. You should see the list all cookies stored for Humble Bundle website. Double-click on the row with name `_simpleauth_sess`
7. Copy the **Value** in the popup window

For a visual guide on steps 4 to 7, [see this picture](safari.jpg).

[hb]: https://www.humblebundle.com
