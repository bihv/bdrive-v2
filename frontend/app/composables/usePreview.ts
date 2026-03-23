// Preview utility composable
// Provides helpers for file preview functionality

export interface PreviewData {
    url: string
    mime_type: string
    name: string
    size: number
    updated_at: string
}

export type PreviewType = 'image' | 'video' | 'audio' | 'pdf' | 'text' | 'office' | 'unknown'

const IMAGE_EXTENSIONS = ['jpg', 'jpeg', 'png', 'gif', 'bmp', 'webp', 'svg', 'ico', 'tiff', 'tif', 'avif']
const VIDEO_EXTENSIONS = ['mp4', 'webm', 'ogg', 'mov', 'avi', 'mkv', 'flv', 'wmv', 'm4v']
const AUDIO_EXTENSIONS = ['mp3', 'wav', 'ogg', 'aac', 'flac', 'wma', 'm4a', 'opus']
const PDF_EXTENSIONS = ['pdf']

const OFFICE_WORD_EXTENSIONS = ['doc', 'docx', 'odt', 'rtf', 'dotx', 'ott']
const OFFICE_CELL_EXTENSIONS = ['xls', 'xlsx', 'ods', 'csv', 'xltx', 'ots']
const OFFICE_SLIDE_EXTENSIONS = ['ppt', 'pptx', 'odp', 'ppsx', 'potx', 'otp']
const OFFICE_EXTENSIONS = [...OFFICE_WORD_EXTENSIONS, ...OFFICE_CELL_EXTENSIONS, ...OFFICE_SLIDE_EXTENSIONS]

const TEXT_EXTENSIONS = [
    'txt', 'md', 'json', 'xml', 'log',
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

// Extension → PreviewType lookup
const EXTENSION_TYPE_MAP = new Map<string, PreviewType>(
    [
        ...IMAGE_EXTENSIONS.map(e => [e, 'image'] as const),
        ...VIDEO_EXTENSIONS.map(e => [e, 'video'] as const),
        ...AUDIO_EXTENSIONS.map(e => [e, 'audio'] as const),
        ...PDF_EXTENSIONS.map(e => [e, 'pdf'] as const),
        ...OFFICE_EXTENSIONS.map(e => [e, 'office'] as const),
        ...TEXT_EXTENSIONS.map(e => [e, 'text'] as const),
    ]
)

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
    md: 'markdown', mdx: 'markdown',
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

    /**
     * Determine the preview type based on file extension
     */
    function getPreviewType(fileName?: string): PreviewType {
        if (!fileName) return 'unknown'
        const ext = getFileExtension(fileName)
        return EXTENSION_TYPE_MAP.get(ext) || 'unknown'
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

    /**
     * Open file preview - either in current tab (normal files) or new tab (Office files)
     */
    function openPreview(item: { id: string; name: string }) {
        const router = useRouter()
        if (isOfficeFile(item.name)) {
            // Office files → new tab
            const url = router.resolve({ path: '/office', query: { id: item.id } })
            window.open(url.href, '_blank')
        } else {
            // Other files → trigger modal
            previewItemId.value = item.id
        }
    }

    return {
        previewItemId,
        getPreviewData,
        isOfficeFile,
        getPreviewType,
        getOfficeDocumentType,
        getFileExtension,
        getMonacoLanguage,
        getOnlyOfficeUrl,
        openPreview,
    }
}
