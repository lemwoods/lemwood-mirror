// SEO meta 统一写入：title + description + og/twitter 标签。
// 各视图只需声明 { title, description }，不再各自手写 DOM 查询；
// fullTitle: true 表示 title 已含站名后缀，跳过拼接。

const ensureMeta = (attr, key) => {
  const selector = `meta[${attr}="${key}"]`
  let el = document.querySelector(selector)
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  return el
}

export function useSeoMeta({ title, description, fullTitle = false }, nameFull) {
  return () => {
    const finalTitle = fullTitle || !nameFull ? title : `${title} - ${nameFull}`
    document.title = finalTitle

    if (description) {
      ensureMeta('name', 'description').setAttribute('content', description)
      ensureMeta('property', 'og:title').setAttribute('content', finalTitle)
      ensureMeta('property', 'og:description').setAttribute('content', description)
      ensureMeta('property', 'twitter:title').setAttribute('content', finalTitle)
      ensureMeta('property', 'twitter:description').setAttribute('content', description)
    }
  }
}
