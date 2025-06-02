from __future__ import annotations
from typing import get_origin, get_args, Union, List, Any
from temporalio import workflow
with workflow.unsafe.imports_passed_through():
    from pydantic import BaseModel
def model_to_template(
    model_cls: type[BaseModel],
    indent: int = 0,
    *,
    bullet_lists: bool = True,       # ← теперь True по-умолчанию
    show_optional: bool = True,
) -> str:
    IND = "  "                       # базовый шаг отступа (2 пробела)
    lines: list[str] = []

    # совместимость v1/v2
    try:
        fields = model_cls.model_fields           # v2
    except AttributeError:
        fields = model_cls.__fields__             # v1  # type: ignore[attr-defined]

    for fname, finfo in fields.items():
        desc = getattr(finfo, "description", "") or ""
        outer_t: Any = getattr(finfo, "annotation",
                               getattr(finfo, "type_", None))

        # скрываем Optional, если нужно
        if (not show_optional and
            get_origin(outer_t) is Union and type(None) in get_args(outer_t)):
            continue

        prefix = IND * indent
        main_line = f"{prefix}{fname}: {desc}".rstrip()

        # ── список моделей ───────────────────────────────────────────────
        if get_origin(outer_t) in (list, List):
            inner_t = get_args(outer_t)[0]
            if isinstance(inner_t, type) and issubclass(inner_t, BaseModel):
                # строим вложенный блок с отступом +2
                nested_block = model_to_template(
                    inner_t, indent + 2,
                    bullet_lists=bullet_lists,
                    show_optional=show_optional
                )
                inner_lines = nested_block.splitlines()

                if bullet_lists and inner_lines:
                    # первая строка получит «- »
                    bullet = f"{IND * (indent + 1)}- {inner_lines[0].lstrip()}"
                    # остальные строки остаются как есть
                    nested_block = "\n".join([bullet] + inner_lines[1:])

                main_line = f"{main_line}\n{nested_block}"

        # ── вложенная модель ──────────────────────────────────────────────
        elif isinstance(outer_t, type) and issubclass(outer_t, BaseModel):
            nested_block = model_to_template(
                outer_t, indent + 1,
                bullet_lists=bullet_lists,
                show_optional=show_optional
            )
            main_line = f"{main_line}\n{nested_block}"

        lines.append(main_line)

    return "\n".join(lines)
