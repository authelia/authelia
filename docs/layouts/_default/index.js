var docs = [
{{ range $index, $page := (where .Site.Pages "Section" "configuration") -}}
  {
    id: {{ $index }},
    title: "{{ .Title }}",
    description: "{{ .Params.description }}",
    href: "{{ .URL | relURL }}"
  },
{{ end -}}
];
