// Release notes keep their published URLs and the flat file layout used by
// the release workflow. Only the navigation groups them by major/minor version.
export function groupReleases(pageMap) {
  return pageMap.map(item => {
    if (item.name !== 'changelog' || !item.children) return item

    const children = []
    const groups = new Map()
    const meta = item.children.find(child => 'data' in child)?.data || {}

    for (const child of item.children) {
      const version = /^v(\d+)\.(\d+)\.\d+$/.exec(child.name || '')
      if (!version) {
        children.push(child)
        continue
      }

      const name = `v${version[1]}.${version[2]}.x`
      if (!groups.has(name)) {
        groups.set(name, {
          name,
          title: name,
          route: `${item.route}/${name}`,
          children: [{ data: {} }],
        })
      }
      const group = groups.get(name)
      group.children[0].data[child.name] = meta[child.name] || {}
      group.children.push(child)
    }

    const newestFirst = (a, b) => b.name.localeCompare(a.name, 'en', { numeric: true })
    const versions = [...groups.values()].sort(newestFirst)
    for (const group of versions) {
      const [metadata, ...releases] = group.children
      group.children = [metadata, ...releases.sort(newestFirst)]
    }
    return { ...item, children: [...children, ...versions] }
  })
}
