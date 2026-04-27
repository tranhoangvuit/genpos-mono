import { Link, useRouterState } from '@tanstack/react-router'
import {
  BarChart3,
  ChevronRight,
  LayoutDashboard,
  Package,
  Puzzle,
  Settings,
  Users,
  Warehouse,
} from 'lucide-react'
import { useTranslation } from 'react-i18next'

import { useAuthStore } from '@/shared/auth/store'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarMenuSub,
  SidebarMenuSubButton,
  SidebarMenuSubItem,
} from '@/shared/ui/sidebar'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/shared/ui/collapsible'

type NavItem = {
  labelKey: string
  icon: React.ComponentType<{ className?: string }>
  to?: string
  children?: { labelKey: string; to: string }[]
}

const NAV_ITEMS: NavItem[] = [
  {
    labelKey: 'nav.home',
    icon: LayoutDashboard,
    to: '/$subdomain/dashboard',
  },
  {
    labelKey: 'nav.products',
    icon: Package,
    children: [
      { labelKey: 'nav.allProducts', to: '/$subdomain/products' },
      { labelKey: 'nav.categories', to: '/$subdomain/products/categories' },
    ],
  },
  {
    labelKey: 'nav.inventory',
    icon: Warehouse,
    children: [
      { labelKey: 'nav.stocks', to: '/$subdomain/inventory/stocks' },
      { labelKey: 'nav.suppliers', to: '/$subdomain/inventory/suppliers' },
      { labelKey: 'nav.purchaseOrders', to: '/$subdomain/inventory/purchase-orders' },
      // { labelKey: 'nav.stockIns', to: '/$subdomain/inventory/stock-ins' },
      { labelKey: 'nav.stockTakes', to: '/$subdomain/inventory/stock-takes' },
    ],
  },
  {
    labelKey: 'nav.customers',
    icon: Users,
    children: [
      { labelKey: 'nav.allCustomers', to: '/$subdomain/customers' },
      { labelKey: 'nav.customerGroups', to: '/$subdomain/customers/groups' },
    ],
  },
  {
    labelKey: 'nav.reports',
    icon: BarChart3,
    children: [
      { labelKey: 'nav.dailySalesReport', to: '/$subdomain/daily-sales-report' },
    ],
  },
  {
    labelKey: 'nav.addons',
    icon: Puzzle,
    to: '/$subdomain/addons',
  },
  {
    labelKey: 'nav.settings',
    icon: Settings,
    children: [
      { labelKey: 'nav.stores', to: '/$subdomain/settings/stores' },
      { labelKey: 'nav.payments', to: '/$subdomain/settings/payments' },
      { labelKey: 'nav.taxes', to: '/$subdomain/settings/taxes' },
      { labelKey: 'nav.staffs', to: '/$subdomain/settings/staffs' },
    ],
  },
]

export function AppSidebar() {
  const { t } = useTranslation()
  const pathname = useRouterState({ select: (s) => s.location.pathname })
  const subdomain = useAuthStore((s) => s.user?.orgSlug ?? '')

  const resolve = (to: string) => to.replace('$subdomain', subdomain)

  return (
    <Sidebar>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV_ITEMS.map((item) => {
                if (!item.children) {
                  const resolved = resolve(item.to!)
                  const active = pathname === resolved
                  return (
                    <SidebarMenuItem key={item.labelKey}>
                      <SidebarMenuButton
                        asChild
                        isActive={active}
                        tooltip={t(item.labelKey)}
                      >
                        <Link to={item.to!} params={{ subdomain }}>
                          <item.icon />
                          <span>{t(item.labelKey)}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                }

                const isGroupActive = item.children.some((c) =>
                  pathname.startsWith(resolve(c.to))
                )

                return (
                  <Collapsible
                    key={item.labelKey}
                    asChild
                    defaultOpen={isGroupActive}
                    className="group/collapsible"
                  >
                    <SidebarMenuItem>
                      <CollapsibleTrigger asChild>
                        <SidebarMenuButton tooltip={t(item.labelKey)}>
                          <item.icon />
                          <span>{t(item.labelKey)}</span>
                          <ChevronRight className="ml-auto transition-transform duration-200 group-data-[state=open]/collapsible:rotate-90" />
                        </SidebarMenuButton>
                      </CollapsibleTrigger>
                      <CollapsibleContent>
                        <SidebarMenuSub>
                          {item.children.map((child) => {
                            const resolved = resolve(child.to)
                            return (
                              <SidebarMenuSubItem key={child.to}>
                                <SidebarMenuSubButton
                                  asChild
                                  isActive={pathname === resolved}
                                >
                                  <Link to={child.to} params={{ subdomain }}>
                                    <span>{t(child.labelKey)}</span>
                                  </Link>
                                </SidebarMenuSubButton>
                              </SidebarMenuSubItem>
                            )
                          })}
                        </SidebarMenuSub>
                      </CollapsibleContent>
                    </SidebarMenuItem>
                  </Collapsible>
                )
              })}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  )
}
