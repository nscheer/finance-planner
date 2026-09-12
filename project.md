# Financial planner

Create a simple financial planner.

## Goal and usage

The goal is to enter spendings and categorize them.
Categories have to be added first, before spendings can be entered.
Each spending has a name, an amount in € and a category.
A Category can't be deleted unless it is not used anymore.
A spending can be deleted.

When adding a spending, I want to chose if I pay this by month or yearly.
This is saved per spending in a datastructure.

Similiar for income. An income has an amount in € and a category.
It works similar to the spending.

The main application view is a table that lists all the income and spendings and groups them by category.
This is shown in two distinct blocks, income first, spendings below.

Per entry (no matter of income or spending) it shows how much it is per month and yearly.
This is calculated depending on what the user entered (so the missing value is calculated).
It should be visible if it is a monthly or yearly entry (so what is the master of the entry).

Income and Spending categories should be sortable via drag & drop. This is saved as well.
Entries themselves should be sortable as well via drag & drop - within the respective category.
It should be possible to drag & drop an entry to a different category - there does not need to be any confirmation or modal, but it should be clearly visible,
where the target location of the dragged entry is.

Categories should be collapsible, and at the top there should be a "expand all", "collapse all" function.

On the right there should be a box with some statistics:

- average cost per month
- saldo per month
- saldo per year

## Background information:
What I do is the following: If a spending is per month, I need to send money to it to my bank account.
If it is a yearly spending, I should save the yearly spending divided by 12 per month.
So when the spending happens, I have saved enough money to move it from savings account to bank account.
Of course this only works as a sliding window, but this way it is the easiest.

For this to work nicely the statistics box should contain the amount of money I need to transfer to the banking account monthly (fed from the monthly spendings),
and also what I need to send to the savings account (so 1/12 of each spending that is a yearly spending).

## Implementation
Use wails3 and svelte + typescript for implemeting this. wails3 is already installed, so is npm. You should not need to install anything.
If you need to install anything (npm packages or other stuff is of course ok), ask first.

Make it look nice - the table should have clear fonts and a clear color schema.
Create clean data structures for all the types needed and persist them to disk as JSON. It is best to use a single structures JSON file.

It should be possible to export and import the JSON file.
The JSON file should be stored next to the binary and be called "data.json".
When importing ask if the data should be added or if the current data should be replaced by the import completely.
The JSON file should carry a version number, so that it could be converted, should the structure be changed in the future.

The directory is already an empty git repository.
It is ok to work on the main branch - but structure your work in commits and do commit after stages that make sense.

The window size and the page should be optimized for larger screen in wide mode, but it should be readable if the screen is too wide, so don't stretch it endlessly.
Also there should be a minimum size, before everything gets squished.

Create components and reusable code, add comments where helpful.

Create test cases against the specification in this file.

### Dialogs
Every data entered should be in a modal dialog. When deleting ask for confirmation, also in a modal.
Notifications can be in a modal as well or just as popups, depending on severity.

## General guideline

Read these instruction in full - if there are question, ask them first. When you start to implement, do not ask any question until finished.
