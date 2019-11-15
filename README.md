update new version of language

## Usage

in js file
```js
const VIVNLang = () => import(/*webpackChunkName: "vi-VN" */ '@subiz/lang/widget/vi-VN.json')
const ENUSLang = () => import(/*webpackChunkName: "vi-VN" */ '@subiz/lang/widget/en-US.json')
```

# To update lang
```
edit .po files in dashboard/ or widget/

$ go build
$ ./lang

$ git commit
$ git push

edit package.json

$ npm publish
```
