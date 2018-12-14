.. |br| raw:: html

   <br/>

.. _sql_format:

Extended SQL Format
===================

The key goal of the `sqlr` package is to simplify some of the more tedious
aspects of preparing SQL queries in Go code. One of the most tiresome chores
(at least in the author's opinion) is preparing SQL queries that have a 
large number of columns and placeholders, and matching them against fields
in a Go language struct.

One of the appealing aspects of an ORM is that it will map the inputs
and outputs of an SQL query into the corresponding fields of a Go struct,
but in doing so it creates the SQL as part of the process.

Package `sqlr` takes a different approach. It assumes that the programmer
knows the SQL that they want to execute, they just don't want to mess
around with the tedium of getting the column lists correct and in the 
correct order for the placeholders and struct fields.

To this end, the `sqlr` package knows how to tokenize a string containing
SQL, and it looks for curly braces ``{}``. Anytime it finds them it knows
it has to substitute a list of columns.

.. code-block:: postgres

    -- Examples of "extended" SQL

    select {} from table_name where {} order by {};

    insert into table_name({}) values({});

    update table_name set {} where {};

    delete from table_name where {};

Column lists in each SQL clause
-------------------------------

 +-----------------------------------------------+--------------+---------------------------+
 | SQL Clause                                    | Columns      | Format                    |
 +===============================================+==============+===========================+
 | ``SELECT {}``                                 | All columns  | ``col1,col2,...``         |
 +-----------------------------------------------+--------------+---------------------------+
 | ``SELECT ... FROM ... WHERE {}``              | Primary key  | ``col1=? and col2=? ...`` |
 +-----------------------------------------------+--------------+---------------------------+
 | ``SELECT ... FROM ... WHERE ... ORDER BY {}`` | Primary key  | ``col1,col2,...``         |
 +-----------------------------------------------+--------------+---------------------------+
 | ``UPDATE ... SET {}``                         | Updateable   | ``col1=?,col2=?,...``     |
 +-----------------------------------------------+--------------+---------------------------+
 | ``UPDATE ... SET ... WHERE {}``               | Primary key  | ``col1=? and col2=? ...`` |
 +-----------------------------------------------+--------------+---------------------------+
 | ``INSERT INTO ...({})``                       | Insertable   | ``col1,col2,...``         |
 +-----------------------------------------------+--------------+---------------------------+
 | ``INSERT INTO ... VALUES ({})``               | Insertable   | ``?,?,...``               |
 +-----------------------------------------------+--------------+---------------------------+
 | ``DELETE FROM ... WHERE {}``                  | Primary key  | ``col1=? and col2=? ...`` |
 +-----------------------------------------------+--------------+---------------------------+

The columns that are included in each list are determined from the contents of
the Go struct field tags.

============  ====================================================
Columns       Inclusion criteria from struct field tag 
============  ====================================================
All columns   Does not contain ``-``                               
Primary key   Contains ``primary key``                             
Updateable    Contains neither ``primary key`` nor ``autoincrement``
Insertable    Does not contain `autoincrement`                   
============  ====================================================

The ``{}`` can be modified with a small set of simple modifiers

+---------------+------------------------------------------------------+
| Symbol        | Meaning                                              |
+===============+======================================================+
| ``{}``        | Column list as per the SQL clause                    |
+---------------+------------------------------------------------------+
| ``{alias x}`` | Prefix the column list with ``x``:                   |
|               |                                                      |
|               |  ``x.col1,x.col2,...`` or                            |
|               |                                                      |
|               |  ``x.col1=? and x.col2=? ...``, etc                  |
+---------------+------------------------------------------------------+
| ``{pk}``      | Override column list to contain only primary key     |
|               | columns                                              |
+---------------+------------------------------------------------------+
| ``{all}``     | Override column list to contain all columns          |
+---------------+------------------------------------------------------+

Example of using aliases:

.. code-block:: postgres

    select {alias u}
    from users u
    inner join user_search_terms t on t.user_id = u.id
    where t.search_term like ?
