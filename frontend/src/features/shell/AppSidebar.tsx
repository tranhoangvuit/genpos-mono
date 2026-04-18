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
    to: '/dashboard',
  },
  {
    labelKey: 'nav.products',
    icon: Package,
    children: [
      { labelKey: 'nav.allProducts', to: '/products' },
      { labelKey: 'nav.categories', to: '/products/categories' },
    ],
  },
  {
    labelKey: 'nav.inventory',
    icon: Warehouse,
    children: [
      { labelKey: 'nav.suppliers', to: '/inventory/suppliers' },
      { labelKey: 'nav.purchaseOrders', to: '/inventory/purchase-orders' },
      { labelKey: 'nav.stockTakes', to: '/inventory/stock-takes' },
    ],
  },
  {
    labelKey: 'nav.customers',
    icon: Users,
    children: [
      { labelKey: 'nav.allCustomers', to: '/customers' },
      { labelKey: 'nav.customerGroups', to: '/customers/groups' },
    ],
  },
  {
    labelKey: 'nav.reports',
    icon: BarChart3,
    children: [
      { labelKey: 'nav.dailySalesReport', to: '/daily-sales-report' },
    ],
  },
  {
    labelKey: 'nav.addons',
    icon: Puzzle,
    to: '/addons',
  },
  {
    labelKey: 'nav.settings',
    icon: Settings,
    children: [
      { labelKey: 'nav.stores', to: '/settings/stores' },
      { labelKey: 'nav.payments', to: '/settings/payments' },
      { labelKey: 'nav.taxes', to: '/settings/taxes' },
      { labelKey: 'nav.members', to: '/settings/members' },
    ],
  },
]

export function AppSidebar() {
  const { t } = useTranslation()
  const pathname = useRouterState({ select: (s) => s.location.pathname })

  return (
    <Sidebar>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {NAV_ITEMS.map((item) => {
                if (!item.children) {
                  const active = pathname === item.to
                  return (
                    <SidebarMenuItem key={item.labelKey}>
                      <SidebarMenuButton
                        asChild
                        isActive={active}
                        tooltip={t(item.labelKey)}
                      >
                        <Link to={item.to!}>
                          <item.icon />
                          <span>{t(item.labelKey)}</span>
                        </Link>
                      </SidebarMenuButton>
                    </SidebarMenuItem>
                  )
                }

                const isGroupActive = item.children.some((c) =>
                  pathname.startsWith(c.to)
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
                          {item.children.map((child) => (
                            <SidebarMenuSubItem key={child.to}>
                              <SidebarMenuSubButton
                                asChild
                                isActive={pathname === child.to}
                              >
                                <Link to={child.to}>
                                  <span>{t(child.labelKey)}</span>
                                </Link>
                              </SidebarMenuSubButton>
                            </SidebarMenuSubItem>
                          ))}
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
