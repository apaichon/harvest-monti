import type { Category, Item, Promo } from "@monti/shared";

export const categories: Category[] = [
  { id: "all", name: "All" },
  { id: "hot", name: "Hot Drinks" },
  { id: "cold", name: "Cold Drinks" },
  { id: "pizza", name: "Pizza" },
  { id: "pasta", name: "Pasta" },
  { id: "burger", name: "Burger" },
  { id: "salad", name: "Salads" },
];

const pizzaImg =
  "https://images.unsplash.com/photo-1565299624946-b28f40a0ae38?w=480&auto=format";
const saladImg =
  "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=480&auto=format";
const pastaImg =
  "https://images.unsplash.com/photo-1551183053-bf91a1d81141?w=480&auto=format";
const burgerImg =
  "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=480&auto=format";
const latteImg =
  "https://images.unsplash.com/photo-1517701550927-30cf4ba1dba5?w=480&auto=format";
const truffleImg =
  "https://images.unsplash.com/photo-1473093295043-cdd812d0e601?w=480&auto=format";

export const items: Item[] = [
  {
    id: "truffle-pasta-001",
    sku: "truffle-pasta-001",
    categoryId: "pasta",
    name: "Truffle Pasta",
    description: "Hand-rolled tagliatelle, black truffle cream, parmesan.",
    imageUrl: truffleImg,
    price: 320,
    isBestSeller: true,
    inStock: true,
    modifierGroups: [
      {
        id: "add-side",
        name: "Add Side",
        kind: "single",
        required: true,
        options: [
          { id: "salad", name: "Salad", priceDelta: 0 },
          { id: "soup", name: "Soup", priceDelta: 0 },
          { id: "bread", name: "Bread", priceDelta: 0 },
        ],
      },
    ],
  },
  {
    id: "margherita-001",
    sku: "margherita-001",
    categoryId: "pizza",
    name: "Margherita Pizza",
    description: "Tomato, mozzarella, fresh basil. Wood-fired 12 inch base.",
    imageUrl: pizzaImg,
    price: 240,
    isBestSeller: true,
    inStock: true,
    modifierGroups: [
      {
        id: "add-side",
        name: "Add Side",
        kind: "single",
        required: true,
        options: [
          { id: "salad", name: "Salad", priceDelta: 0 },
          { id: "soup", name: "Soup", priceDelta: 0 },
          { id: "bread", name: "Bread", priceDelta: 0 },
        ],
      },
      {
        id: "toppings",
        name: "Toppings",
        kind: "multi",
        required: false,
        options: [
          { id: "mushrooms", name: "Extra Mushrooms", priceDelta: 30 },
          { id: "olives", name: "Olives", priceDelta: 20 },
          { id: "anchovies", name: "Anchovies", priceDelta: 40 },
        ],
      },
    ],
  },
  {
    id: "caesar-001",
    sku: "caesar-001",
    categoryId: "salad",
    name: "Caesar Salad",
    description: "Crisp romaine, parmesan, garlic croutons.",
    imageUrl: saladImg,
    price: 180,
    isNew: true,
    inStock: true,
  },
  {
    id: "carbonara-001",
    sku: "carbonara-001",
    categoryId: "pasta",
    name: "Penne Carbonara",
    description: "Guanciale, egg yolk, pecorino.",
    imageUrl: pastaImg,
    price: 220,
    inStock: true,
  },
  {
    id: "cheeseburger-001",
    sku: "cheeseburger-001",
    categoryId: "burger",
    name: "Cheeseburger",
    description: "200g chuck patty, aged cheddar, brioche bun.",
    imageUrl: burgerImg,
    price: 260,
    inStock: true,
  },
  {
    id: "iced-latte-001",
    sku: "iced-latte-001",
    categoryId: "cold",
    name: "Iced Latte",
    description: "Single-origin espresso, cold milk, ice.",
    imageUrl: latteImg,
    price: 120,
    inStock: true,
  },
];

export const promos: Promo[] = [
  {
    code: "LUNCH20",
    label: "Lunch 20% off",
    discountKind: "percent",
    discountValue: 20,
    validUntil: "14:00",
  },
  {
    code: "FAMILY200",
    label: "฿200 off ฿1000+",
    discountKind: "fixed",
    discountValue: 200,
    minSubtotal: 1000,
  },
];
