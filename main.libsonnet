{
  File(name, content=""):: {
    local file = self,

    apiVersion: "jsonnet-filer.zeet.co/v1alpha1",
    kind: "File",
    metadata: {
      name: name,
    },
    content: content,
    encodingStrategy: "yaml",
  }
}
