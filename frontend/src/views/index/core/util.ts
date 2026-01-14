import type {IWorkbookData} from "@univerjs/presets";
import type {IFUniverSheetsMixin} from "@univerjs/sheets/lib/types/facade/f-univer";

export const importWorkbookJSONData = (univerAPI:IFUniverSheetsMixin,data:Partial<IWorkbookData>) => {
    univerAPI.createWorkbook(data)
}

/**
 * 可拖拽元素，支持排除内部节点 & 视口边界限制
 * @param el 要拖拽的元素
 * @param excludeSelector 内部不触发拖拽的选择器
 */
export const useDraggable = (el: HTMLElement | null, excludeSelector?: string) => {
    if (!el) return;

    let offsetX = 0;
    let offsetY = 0;
    let dragging = false;
    let moved = false; // 是否移动过

    const onMouseDown = (e: MouseEvent) => {
        // 排除内部元素
        if (excludeSelector && (e.target as HTMLElement).closest(excludeSelector)) {
            return;
        }

        dragging = true;
        moved = false;
        const rect = el.getBoundingClientRect();
        offsetX = e.clientX - rect.left;
        offsetY = e.clientY - rect.top;

        document.addEventListener("mousemove", onMouseMove);
        document.addEventListener("mouseup", onMouseUp);

        e.preventDefault(); // 防止选中文本
    };

    const onMouseMove = (e: MouseEvent) => {
        if (!dragging) return;

        moved = true; // 已经移动过
        const viewportWidth = window.innerWidth;
        const viewportHeight = window.innerHeight;

        const elWidth = el.offsetWidth;
        const elHeight = el.offsetHeight;

        let left = e.clientX - offsetX;
        let top = e.clientY - offsetY;

        left = Math.max(0, Math.min(left, viewportWidth - elWidth));
        top = Math.max(0, Math.min(top, viewportHeight - elHeight));

        el.style.left = left + "px";
        el.style.top = top + "px";
        el.style.right = "auto"; // 取消右对齐
    };

    const onMouseUp = () => {
        if (dragging && moved) {
            // 拖动过，阻止 click 事件冒泡
            const preventClick = (clickEvent: MouseEvent) => {
                clickEvent.stopImmediatePropagation();
                clickEvent.preventDefault();
                el.removeEventListener("click", preventClick, true);
            };
            el.addEventListener("click", preventClick, true);
        }

        dragging = false;
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
    };

    el.addEventListener("mousedown", onMouseDown);

    // 返回销毁函数
    return () => {
        el.removeEventListener("mousedown", onMouseDown);
        document.removeEventListener("mousemove", onMouseMove);
        document.removeEventListener("mouseup", onMouseUp);
    };
};
