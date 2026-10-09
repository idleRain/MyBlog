import type { DictItem, DictType } from '@myblog/api/modules/dict/types'
import { beforeEach, describe, expect, it, vi } from 'vitest'

// 接口与提示均以替身注入，测试只验证页面状态的编排逻辑，不触达网络与界面。
const apiMock = {
  adminListTypes: vi.fn(),
  adminListItems: vi.fn(),
  adminCreateType: vi.fn(),
  adminUpdateType: vi.fn(),
  adminDeleteType: vi.fn(),
  adminCreateItem: vi.fn(),
  adminUpdateItem: vi.fn(),
  adminDeleteItem: vi.fn()
}

const toastMock = {
  success: vi.fn(),
  error: vi.fn()
}

vi.mock('$lib/api', () => ({ DictAPI: apiMock }))
vi.mock('svelte-sonner', () => ({ toast: toastMock }))

const { DICT_PAGE_SIZE, emptyDictPageData, dictPageState } =
  await import('./dict-page-state.svelte')

// 后端统一响应的成功业务码。
const SUCCESS_CODE = 200

// 构造一个字典类型替身，字段取值只覆盖被测逻辑关心的部分。
function makeType(id: number, name = `类型${id}`): DictType {
  return { id, name, code: `type_${id}` } as unknown as DictType
}

// 构造一个字典项替身。
function makeItem(id: number, typeId: number): DictItem {
  return { id, typeId, label: `项${id}`, value: `item_${id}` } as unknown as DictItem
}

// 重置替身调用记录与页面状态，避免用例之间互相污染。
function resetState() {
  vi.clearAllMocks()
  dictPageState.initialize(emptyDictPageData())
}

describe('字典页状态：初始化', () => {
  beforeEach(resetState)

  it('用 load 结果填充全部字段', () => {
    const types = [makeType(1), makeType(2)]
    const items = [makeItem(10, 2)]

    dictPageState.initialize({
      types,
      typesTotal: 2,
      selectedTypeId: 2,
      items,
      itemsTotal: 1
    })

    expect(dictPageState.types).toEqual(types)
    expect(dictPageState.typesTotal).toBe(2)
    expect(dictPageState.selectedTypeId).toBe(2)
    expect(dictPageState.items).toEqual(items)
    expect(dictPageState.itemsTotal).toBe(1)
  })

  it('selectedType 按当前选中 id 解析，未选中时返回 null', () => {
    const types = [makeType(1), makeType(2)]
    dictPageState.initialize({
      types,
      typesTotal: 2,
      selectedTypeId: null,
      items: [],
      itemsTotal: 0
    })

    expect(dictPageState.selectedType).toBeNull()

    dictPageState.selectedTypeId = 2
    expect(dictPageState.selectedType?.id).toBe(2)
  })

  it('空页面数据给出全零初值', () => {
    const empty = emptyDictPageData()

    expect(empty.types).toEqual([])
    expect(empty.typesTotal).toBe(0)
    expect(empty.selectedTypeId).toBeNull()
    expect(empty.items).toEqual([])
    expect(empty.itemsTotal).toBe(0)
  })
})

describe('字典页状态：加载类型列表', () => {
  beforeEach(resetState)

  it('请求页码固定为首页并带单页容量', async () => {
    apiMock.adminListTypes.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { types: [makeType(1)], total: 1 }
    })
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [], total: 0 }
    })

    await dictPageState.loadTypes()

    expect(apiMock.adminListTypes).toHaveBeenCalledWith({ page: 1, pageSize: DICT_PAGE_SIZE })
  })

  it('选中类型仍在新列表中时保持不动，不重新加载字典项', async () => {
    dictPageState.initialize({
      types: [makeType(1), makeType(2)],
      typesTotal: 2,
      selectedTypeId: 2,
      items: [makeItem(20, 2)],
      itemsTotal: 1
    })
    apiMock.adminListTypes.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { types: [makeType(1), makeType(2)], total: 2 }
    })

    await dictPageState.loadTypes()

    expect(dictPageState.selectedTypeId).toBe(2)
    expect(dictPageState.items).toHaveLength(1)
    expect(apiMock.adminListItems).not.toHaveBeenCalled()
  })

  it('选中类型被删除后回退到首个类型并加载其字典项', async () => {
    dictPageState.initialize({
      types: [makeType(1), makeType(9)],
      typesTotal: 2,
      selectedTypeId: 9,
      items: [makeItem(90, 9)],
      itemsTotal: 1
    })
    apiMock.adminListTypes.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { types: [makeType(1), makeType(2)], total: 2 }
    })
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [makeItem(11, 1)], total: 1 }
    })

    await dictPageState.loadTypes()

    expect(dictPageState.selectedTypeId).toBe(1)
    expect(apiMock.adminListItems).toHaveBeenCalledTimes(1)
    expect(dictPageState.items).toHaveLength(1)
  })

  it('类型列表为空时清空选中与字典项', async () => {
    dictPageState.initialize({
      types: [makeType(1)],
      typesTotal: 1,
      selectedTypeId: 1,
      items: [makeItem(10, 1)],
      itemsTotal: 1
    })
    apiMock.adminListTypes.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { types: [], total: 0 }
    })

    await dictPageState.loadTypes()

    expect(dictPageState.selectedTypeId).toBeNull()
    expect(dictPageState.items).toEqual([])
    expect(dictPageState.itemsTotal).toBe(0)
    expect(apiMock.adminListItems).not.toHaveBeenCalled()
  })

  it('业务失败时提示后端消息且不改动既有列表', async () => {
    dictPageState.initialize({
      types: [makeType(1)],
      typesTotal: 1,
      selectedTypeId: 1,
      items: [],
      itemsTotal: 0
    })
    apiMock.adminListTypes.mockResolvedValue({ code: 500, message: '服务端拒绝' })

    await dictPageState.loadTypes()

    expect(toastMock.error).toHaveBeenCalledWith('服务端拒绝')
    expect(dictPageState.types).toHaveLength(1)
  })

  it('网络异常时提示统一文案，加载标记在结束后复位', async () => {
    apiMock.adminListTypes.mockRejectedValue(new Error('connection reset'))

    await dictPageState.loadTypes()

    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(dictPageState.isLoadingTypes).toBe(false)
  })
})

describe('字典页状态：选中与字典项加载', () => {
  beforeEach(resetState)

  it('重复选中当前类型且不在加载中时跳过重复请求', async () => {
    dictPageState.initialize({
      types: [makeType(1)],
      typesTotal: 1,
      selectedTypeId: 1,
      items: [],
      itemsTotal: 0
    })

    await dictPageState.selectType(1)

    expect(apiMock.adminListItems).not.toHaveBeenCalled()
  })

  it('未选中类型时直接清空字典项，不发起请求', async () => {
    await dictPageState.loadItems()

    expect(apiMock.adminListItems).not.toHaveBeenCalled()
    expect(dictPageState.items).toEqual([])
    expect(dictPageState.itemsTotal).toBe(0)
  })

  it('加载字典项时携带当前选中类型与单页容量', async () => {
    dictPageState.selectedTypeId = 7
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [makeItem(70, 7)], total: 1 }
    })

    await dictPageState.loadItems()

    expect(apiMock.adminListItems).toHaveBeenCalledWith({
      typeId: 7,
      page: 1,
      pageSize: DICT_PAGE_SIZE
    })
    expect(dictPageState.itemsTotal).toBe(1)
  })

  it('响应缺少 data 时按失败处理并提示', async () => {
    dictPageState.selectedTypeId = 7
    apiMock.adminListItems.mockResolvedValue({ code: SUCCESS_CODE, message: '数据缺失' })

    await dictPageState.loadItems()

    expect(toastMock.error).toHaveBeenCalledWith('数据缺失')
    expect(dictPageState.items).toEqual([])
  })
})

describe('字典页状态：写操作后的刷新', () => {
  beforeEach(resetState)

  it('创建类型成功后刷新类型列表并返回成功', async () => {
    apiMock.adminCreateType.mockResolvedValue({ code: SUCCESS_CODE })
    apiMock.adminListTypes.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { types: [makeType(1)], total: 1 }
    })
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [], total: 0 }
    })

    const result = await dictPageState.createType({ name: '新类型', code: 'new_type' })

    expect(result).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('字典类型创建成功')
    expect(apiMock.adminListTypes).toHaveBeenCalledTimes(1)
  })

  it('创建类型失败时不刷新列表并返回失败', async () => {
    apiMock.adminCreateType.mockResolvedValue({ code: 400, message: '编码重复' })

    const result = await dictPageState.createType({ name: '新类型', code: 'new_type' })

    expect(result).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('编码重复')
    expect(apiMock.adminListTypes).not.toHaveBeenCalled()
  })

  it('删除类型成功后清空选中与字典项再刷新', async () => {
    dictPageState.initialize({
      types: [makeType(1)],
      typesTotal: 1,
      selectedTypeId: 1,
      items: [makeItem(10, 1)],
      itemsTotal: 1
    })
    apiMock.adminDeleteType.mockResolvedValue({ code: SUCCESS_CODE })
    apiMock.adminListTypes.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { types: [makeType(2)], total: 1 }
    })
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [], total: 0 }
    })

    const result = await dictPageState.removeType(1)

    expect(result).toBe(true)
    expect(toastMock.success).toHaveBeenCalledWith('字典类型删除成功')
    // 删除时先清空选中，刷新拿到新列表后按「选中缺失则回退首个」的规则重新选中。
    expect(dictPageState.selectedTypeId).toBe(2)
    expect(apiMock.adminListTypes).toHaveBeenCalledTimes(1)
  })

  it('创建字典项成功后刷新当前类型的字典项', async () => {
    dictPageState.selectedTypeId = 3
    apiMock.adminCreateItem.mockResolvedValue({ code: SUCCESS_CODE })
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [makeItem(30, 3)], total: 1 }
    })

    const result = await dictPageState.createItem({ typeId: 3, label: '新项', value: 'new_item' })

    expect(result).toBe(true)
    expect(apiMock.adminListItems).toHaveBeenCalledWith({
      typeId: 3,
      page: 1,
      pageSize: DICT_PAGE_SIZE
    })
  })

  it('更新字典项失败时不刷新并返回失败', async () => {
    // 选中类型必须先就位，否则刷新会被 loadItems 的守卫拦截而无法证明「未发起刷新」。
    dictPageState.selectedTypeId = 5
    apiMock.adminUpdateItem.mockRejectedValue(new Error('timeout'))

    // 更新请求刻意不含 value，字典项值不可变更，故此处只改标签。
    const result = await dictPageState.updateItem({ id: 1, label: '改后的标签' })

    expect(result).toBe(false)
    expect(toastMock.error).toHaveBeenCalledWith('网络错误，请稍后重试')
    expect(apiMock.adminListItems).not.toHaveBeenCalled()
  })

  it('删除字典项成功后刷新当前类型的字典项', async () => {
    dictPageState.selectedTypeId = 4
    apiMock.adminDeleteItem.mockResolvedValue({ code: SUCCESS_CODE })
    apiMock.adminListItems.mockResolvedValue({
      code: SUCCESS_CODE,
      data: { items: [], total: 0 }
    })

    const result = await dictPageState.removeItem(40)

    expect(result).toBe(true)
    expect(apiMock.adminListItems).toHaveBeenCalledTimes(1)
  })
})
