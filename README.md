# miri

A comfy way to download music from Deezer, heavily ~stolen~ inspired by [GoDeez](https://github.com/mathismqn/godeez).

### Variables to configure

Here are the key variables you need to set in `.env`:

1. `ARL_COOKIE`
* **What is it?**: The `arl_cookie` is a session cookie used for authentication with Deezer. Without this cookie, the downloader cannot access your account to retrieve playlists, albums, or songs.
* **How to retrieve it**:
	1.	Open your browser and log in to your Deezer account.
	2.	Open the Developer Tools (right-click on the page and select “Inspect” or press F12).
	3.	Navigate to the Application tab (in Chrome/Edge) or Storage tab (in Firefox).
	4.	In the left panel, look for Cookies and select `https://www.deezer.com`.
	5.	Find the arl cookie and copy its value.

2. `SECRET_KEY`
* **What is it?**: The `secret_key` is a cryptographic value used to decrypt Deezer’s media files.
* **How to retrieve it?**: While we cannot provide the specific secret_key in this documentation, it can be found online through various sources or developer communities that focus on Deezer.

## License

miri is provided under the MIT License.

---

> ⚠️ This tool is provided for educational and personal use only. Please ensure your usage complies with Deezer’s Terms of Service.
