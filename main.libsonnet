{
  File(name, content=""):: {
    local file = self,

    apiVersion: "jsonnet-filer.zeet.co/v1alpha1",
    kind: "File",
    metadata: {
      name: name,
    },
    content: content,

    local encodingStrategies = {
      json: std.manifestJsonEx(file.content, "  "),
      yaml: std.manifestYamlDoc(file.content, indent_array_in_object=false, quote_keys=false),
      // TODO finish this and write tests
      // yamlStream: std.manifestYamlStream(file.content, indent_array_in_object=false, quote_keys=false),
    },

    encodingStrategy: "yaml",
    assert std.objectHas(encodingStrategies, self.encodingStrategy) :
      "unsupported encoding strategy: %s, valid encoding strategies are %s" % [self.encodingStrategy, encodingStrategies],

    contentEncoded: encodingStrategies[self.encodingStrategy],
  }
}
