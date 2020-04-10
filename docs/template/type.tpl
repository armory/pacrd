{{ define "type" }}

### {{ .Name.Name }} {{ if eq .Kind "Alias" }}(<code>{{.Underlying}}</code> alias)</p>{{ end -}}

{{ with (typeReferences .) }}
(__Appears on:__
{{- $prev := "" -}}
{{- range . -}}
    {{- if $prev -}}, {{ end -}}
    {{ $prev = . }}
    <a href="{{ linkForType . }}">{{ typeDisplayName . }}</a>
{{- end -}}
)
{{ end }}

{{ safe (renderComments .CommentLines) }}

{{ if .Members }}
<table>
    <thead>
        <tr>
            <th>Field</th>
            <th>Description</th>
        </tr>
    </thead>
    <tbody>
        {{ if isExportedType . }}
        <tr>
            <td>
                <code>apiVersion</code></br>
                string</td>
            <td>
                <code>
                    {{apiGroup .}}
                </code>
            </td>
        </tr>
        <tr>
            <td>
                <code>kind</code></br>
                string
            </td>
            <td><code>{{.Name.Name}}</code></td>
        </tr>
        {{ end }}
        {{ template "members" .}}
    </tbody>
</table>
{{ end }}

{{ end }}
