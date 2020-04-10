{{ define "packages" }}
---
layout: post
title: PaCRD CRD Documentation
order: 172
---

{:.no_toc}
* This is a placeholder for an unordered list that will be replaced with ToC. To exclude a header, add {:.no_toc} after it.
{:toc}

{{ range .packages }}
# {{ packageDisplayName . -}}

    {{ with (index .GoPackages 0 )}}
        {{ with .DocComments }}
        <p>
            {{ safe (renderComments .) }}
        </p>
        {{ end }}
    {{ end }}

## Resource Types:
    <ul>
    {{- range (visibleTypes (sortedTypes .Types)) -}}
        {{ if isExportedType . -}}
        <li>
            <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
        </li>
        {{- end }}
    {{- end -}}
    </ul>

    {{ range (visibleTypes (sortedTypes .Types))}}
        {{ template "type" .  }}
    {{ end }}
    <hr/>
{{ end }}

<p><em>
    Generated with <code>gen-crd-api-reference-docs</code>
    {{ with .gitCommit }} on git commit <code>{{ . }}</code>{{end}}.
</em></p>

{{ end }}
