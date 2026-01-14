import { RouteConfig } from "types";
import Layout from "@/layout/index.vue";

const routers: RouteConfig[] = [
    {
        path: "/",
        redirect: "/index",
        meta: {
            sort: 100
        },
        component: Layout,
        children: [
            {
                path: "/redirect",
                name: "redirect",
                meta: {
                    title: "redirect",
                    icon: "zx-1-1",
                    sort: 9,
                    hidden: true
                },
                component: () => import("@/views/redirect/index.vue")
            }
        ]
    }
];

export default routers;
