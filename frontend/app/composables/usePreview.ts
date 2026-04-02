// Preview utility composable
// Provides helpers for file preview functionality

export interface PreviewData {
    url: string
    mime_type: string
    name: string
    size: number
    updated_at: string
}

export type PreviewType = 'image' | 'video' | 'audio' | 'pdf' | 'text' | 'markdown' | 'office' | 'unknown'

const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg', 'ico', 'tiff', 'tif', 'avif']
const VIDEO_EXTENSIONS = ['mp4', 'webm', 'ogg', 'mov', 'avi', 'mkv', 'flv', 'wmv', 'm4v']
const AUDIO_EXTENSIONS = ['mp3', 'wav', 'ogg', 'aac', 'flac', 'wma', 'm4a', 'opus']
const PDF_EXTENSIONS = ['pdf']

const OFFICE_WORD_EXTENSIONS = ['doc', 'docx', 'odt', 'rtf', 'dotx', 'ott']
const OFFICE_CELL_EXTENSIONS = ['xls', 'xlsx', 'ods', 'csv', 'xltx', 'ots']
const OFFICE_SLIDE_EXTENSIONS = ['ppt', 'pptx', 'odp', 'ppsx', 'potx', 'otp']
const OFFICE_EXTENSIONS = [...OFFICE_WORD_EXTENSIONS, ...OFFICE_CELL_EXTENSIONS, ...OFFICE_SLIDE_EXTENSIONS]

const MARKDOWN_EXTENSIONS = ['md', 'mdx']
const TEXT_EXTENSIONS = [
    'txt', 'json', 'xml', 'log',
    'js', 'ts', 'jsx', 'tsx', 'vue', 'svelte',
    'py', 'go', 'rs', 'java', 'c', 'cpp', 'h', 'hpp',
    'css', 'scss', 'less', 'html', 'htm',
    'yaml', 'yml', 'toml', 'ini', 'conf',
    'sh', 'bash', 'zsh', 'fish',
    'sql', 'graphql', 'proto',
    'env', 'gitignore', 'dockerignore',
    'rb', 'php', 'swift', 'kt', 'kts', 'lua', 'r',
    'cls', 'trigger', 'apex',
    'dockerfile', 'makefile',
]

type MediaKind = 'video' | 'audio'

const MARKDOWN_MIME_TYPES = new Set(['text/markdown', 'text/x-markdown'])
const TEXT_MIME_TYPES = new Set([
    'application/json',
    'application/xml',
    'application/javascript',
    'application/typescript',
    'application/x-yaml',
    'application/yaml',
    'application/toml',
    'application/x-sh',
])
const MEDIA_EXTENSION_MIME_CANDIDATES: Record<MediaKind, Record<string, string[]>> = {
    video: {
        mp4: ['video/mp4'],
        webm: ['video/webm'],
        ogg: ['video/ogg'],
        m4v: ['video/mp4'],
        mov: ['video/quicktime'],
    },
    audio: {
        mp3: ['audio/mpeg'],
        wav: ['audio/wav', 'audio/x-wav'],
        ogg: ['audio/ogg'],
        aac: ['audio/aac'],
        m4a: ['audio/mp4', 'audio/x-m4a'],
        opus: ['audio/opus', 'audio/ogg'],
    },
}
const FALLBACK_BROWSER_MEDIA_EXTENSIONS: Record<MediaKind, Set<string>> = {
    video: new Set(['mp4', 'webm', 'ogg', 'm4v', 'mov']),
    audio: new Set(['mp3', 'wav', 'ogg', 'aac', 'm4a', 'opus']),
}

// Monaco Editor language mapping (extension -> Monaco language ID)
const MONACO_LANGUAGE_MAP: Record<string, string> = {
    // Web
    js: 'javascript', jsx: 'javascript', mjs: 'javascript', cjs: 'javascript',
    ts: 'typescript', tsx: 'typescript',
    html: 'html', htm: 'html', vue: 'html', svelte: 'html',
    css: 'css', scss: 'scss', less: 'less',
    json: 'json', jsonc: 'json',
    // Scripting
    py: 'python', pyw: 'python',
    rb: 'ruby',
    php: 'php',
    lua: 'lua',
    r: 'r', R: 'r',
    sh: 'shell', bash: 'shell', zsh: 'shell', fish: 'shell',
    // Systems
    go: 'go',
    rs: 'rust',
    c: 'c', h: 'c',
    cpp: 'cpp', cxx: 'cpp', cc: 'cpp', hpp: 'cpp',
    java: 'java',
    kt: 'kotlin', kts: 'kotlin',
    swift: 'swift',
    cs: 'csharp',
    // Salesforce / Apex
    cls: 'apex', trigger: 'apex', apex: 'apex',
    // Data / Config
    xml: 'xml', xsl: 'xml', xsd: 'xml', svg: 'xml',
    yaml: 'yaml', yml: 'yaml',
    toml: 'ini', ini: 'ini', conf: 'ini', env: 'ini',
    sql: 'sql',
    graphql: 'graphql', gql: 'graphql',
    // Docs
    txt: 'plaintext', log: 'plaintext', gitignore: 'plaintext', dockerignore: 'plaintext',
    csv: 'plaintext',
    // DevOps
    dockerfile: 'dockerfile', makefile: 'plaintext',
    proto: 'protobuf',
}

export function usePreview() {
    const api = useApi()
    const config = useRuntimeConfig()

    /**
     * Fetch preview data (pre-signed URL + metadata) from backend
     */
    async function getPreviewData(itemId: string): Promise<PreviewData> {
        return api.get<PreviewData>(`/api/v1/items/${itemId}/preview`)
    }

    /**
     * Check if file is an Office document (by extension)
     */
    function isOfficeFile(fileName?: string): boolean {
        if (!fileName) return false
        const ext = getFileExtension(fileName)
        return OFFICE_EXTENSIONS.includes(ext)
    }

    function normalizeMimeType(mimeType?: string | null): string {
        return mimeType?.split(';')[0]?.trim()?.toLowerCase() || ''
    }

    function isBrowserPreviewableMedia(kind: MediaKind, fileName?: string, mimeType?: string | null): boolean {
        const ext = fileName ? getFileExtension(fileName) : ''
        const normalizedMimeType = normalizeMimeType(mimeType)
        const fallbackSupported = FALLBACK_BROWSER_MEDIA_EXTENSIONS[kind].has(ext)

        if (!import.meta.client) {
            return fallbackSupported
        }

        const mediaElement = document.createElement(kind)
        const mimeCandidates = [
            normalizedMimeType,
            ...(MEDIA_EXTENSION_MIME_CANDIDATES[kind][ext] || []),
        ].filter(Boolean)

        if (mimeCandidates.length === 0) {
            return fallbackSupported
        }

        return mimeCandidates.some(candidate => {
            const support = mediaElement.canPlayType(candidate)
            return support === 'probably' || support === 'maybe'
        })
    }

    function isTextMimeType(mimeType?: string | null): boolean {
        const normalizedMimeType = normalizeMimeType(mimeType)
        return normalizedMimeType.startsWith('text/') || TEXT_MIME_TYPES.has(normalizedMimeType)
    }

    /**
     * Determine the preview type based on file extension
     */
    function getPreviewType(fileName?: string, mimeType?: string | null): PreviewType {
        const ext = fileName ? getFileExtension(fileName) : ''
        const normalizedMimeType = normalizeMimeType(mimeType)

        if (OFFICE_EXTENSIONS.includes(ext)) return 'office'
        if (PDF_EXTENSIONS.includes(ext) || normalizedMimeType === 'application/pdf') return 'pdf'
        if (MARKDOWN_EXTENSIONS.includes(ext) || MARKDOWN_MIME_TYPES.has(normalizedMimeType)) return 'markdown'
        if (TEXT_EXTENSIONS.includes(ext) || isTextMimeType(normalizedMimeType)) return 'text'
        if (IMAGE_EXTENSIONS.includes(ext) || normalizedMimeType.startsWith('image/')) return 'image'

        if (VIDEO_EXTENSIONS.includes(ext) || normalizedMimeType.startsWith('video/')) {
            return isBrowserPreviewableMedia('video', fileName, normalizedMimeType) ? 'video' : 'unknown'
        }

        if (AUDIO_EXTENSIONS.includes(ext) || normalizedMimeType.startsWith('audio/')) {
            return isBrowserPreviewableMedia('audio', fileName, normalizedMimeType) ? 'audio' : 'unknown'
        }

        return 'unknown'
    }

    /**
     * Get OnlyOffice document type from file extension
     */
    function getOfficeDocumentType(fileName?: string): string {
        const ext = fileName?.split('.').pop()?.toLowerCase() || ''
        if (OFFICE_CELL_EXTENSIONS.includes(ext)) return 'cell'
        if (OFFICE_SLIDE_EXTENSIONS.includes(ext)) return 'slide'
        return 'word'
    }

    /**
     * Get file extension from filename
     */
    function getFileExtension(fileName: string): string {
        return fileName.split('.').pop()?.toLowerCase() || ''
    }

    /**
     * Get Monaco Editor language ID from file extension
     */
    function getMonacoLanguage(fileName: string): string {
        const ext = getFileExtension(fileName)
        return MONACO_LANGUAGE_MAP[ext] || 'plaintext'
    }

    /**
     * Get OnlyOffice server URL
     */
    function getOnlyOfficeUrl(): string {
        return config.public.onlyofficeUrl
    }

    const previewItemId = useState<string | null>('preview-item-id', () => null)
    const previewContext = useState<any[]>('preview-context', () => [])
    const previewForceType = useState<PreviewType | null>('preview-force-type', () => null)

    /**
     * Open file preview - either in current tab (normal files) or new tab (Office files)
     */
    function openPreview(item: { id: string; name: string }, contextItems?: any[]) {
        const router = useRouter()
        
        previewForceType.value = null
        if (contextItems) {
            previewContext.value = contextItems
        } else {
            previewContext.value = [item]
        }
        if (isOfficeFile(item.name)) {
            // Office files → new tab
            const url = router.resolve({ path: '/office', query: { id: item.id } })
            window.open(url.href, '_blank')
        } else {
            // Other files → trigger modal
            previewItemId.value = item.id
        }
    }

    /**
     * Open file preview with a forced viewer type (bypass auto-detection)
     */
    function openPreviewAs(item: { id: string; name: string }, forceType: PreviewType, contextItems?: any[]) {
        const router = useRouter()

        if (contextItems) {
            previewContext.value = contextItems
        } else {
            previewContext.value = [item]
        }

        if (forceType === 'office') {
            // Office → always open in new tab
            const url = router.resolve({ path: '/office', query: { id: item.id } })
            window.open(url.href, '_blank')
        } else {
            previewForceType.value = forceType
            previewItemId.value = item.id
        }
    }

    return {
        previewItemId,
        previewContext,
        previewForceType,
        getPreviewData,
        isOfficeFile,
        getPreviewType,
        getOfficeDocumentType,
        getFileExtension,
        getMonacoLanguage,
        getOnlyOfficeUrl,
        openPreview,
        openPreviewAs,
    }
}
