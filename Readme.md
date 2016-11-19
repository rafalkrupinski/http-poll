HTTP poll
=====

> This software is written with Etsy API in mind, but with hope it could be used with any HTTP resource.

Poll a remote HTTP resources and push the contents to another HTTP service.

You configure a set of tasks, each of which is called once in a defined period of time.


In case of Etsy the results are paged and sorted in the descending order in pages.
    If, for example you want to get a hundred documents starting from id X (no way to tell the API how many documents you want, just the page size. It's only an example)
    you will receive a page of 25 documents starting from the most recent instead of starting from X.

In normal processing, you would get all document starting from X, take the maximum id (or minimum in case it's decreasing) and use it as X in the next iteration.
With paged results that normal processing must be suspended and all the pages must be fetched.

This is how it looks like in steps:

    + := downloaded data
    - := data to download
    X := the last ID from the historical data

    historical data++++++X------ <- get the latest documents
    step 1         ++++++X-----+ <- get page
    step 2         ++++++X----++ <- get next page
    step n         ++++++X++++++ <- get nth page
    replace X      ++++++++++++X <- update X for the next iteration
