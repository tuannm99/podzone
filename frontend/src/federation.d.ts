declare const __MFE_BACKOFFICE__: boolean
declare const __MFE_IAM__: boolean
declare const __MFE_ONBOARDING__: boolean

declare module 'backoffice/TenantHomePage' {
    import type { Component } from 'solid-js'
    const TenantHomePage: Component
    export default TenantHomePage
}

declare module 'backoffice/TenantOrdersPage' {
    import type { Component } from 'solid-js'
    const TenantOrdersPage: Component
    export default TenantOrdersPage
}

declare module 'backoffice/TenantOrderAuditPage' {
    import type { Component } from 'solid-js'
    const TenantOrderAuditPage: Component
    export default TenantOrderAuditPage
}

declare module 'backoffice/TenantOrderFinancePage' {
    import type { Component } from 'solid-js'
    const TenantOrderFinancePage: Component
    export default TenantOrderFinancePage
}

declare module 'backoffice/TenantPartnersPage' {
    import type { Component } from 'solid-js'
    const TenantPartnersPage: Component
    export default TenantPartnersPage
}

declare module 'backoffice/TenantPartnerDetailPage' {
    import type { Component } from 'solid-js'
    const TenantPartnerDetailPage: Component
    export default TenantPartnerDetailPage
}

declare module 'backoffice/TenantProductSetupPage' {
    import type { Component } from 'solid-js'
    const TenantProductSetupPage: Component
    export default TenantProductSetupPage
}

declare module 'iam/AdminIamPage' {
    import type { Component } from 'solid-js'
    const AdminIamPage: Component
    export default AdminIamPage
}

declare module 'onboarding/AdminHomePage' {
    import type { Component } from 'solid-js'
    const AdminHomePage: Component
    export default AdminHomePage
}

declare module 'onboarding/AdminProvisioningPage' {
    import type { Component } from 'solid-js'
    const AdminProvisioningPage: Component
    export default AdminProvisioningPage
}
