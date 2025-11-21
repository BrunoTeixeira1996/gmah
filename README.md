# gmah

Cron job that reads email subs about new houses available in my area, used in gokrazy.

To perform an on demand lookup just use curl `curl -v <ip>:9090/demand`

## Testing

>[!Note]
>This is more a note for myself while testing, however is still useful.

- Go to gmail
- Find the email that is not working
- Click in the right upper corner 3 dots
- Click in "view original" and save the `.eml` file in `testdata` folder
- Then run `python3 /gmah/internal/email/generate_html_from_eml.py`
- A new html file will be generated in the `testdata` folder, this can then be used in the tests